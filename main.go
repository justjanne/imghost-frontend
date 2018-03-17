package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func writeBody(reader io.ReadCloser, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}
	return out.Close()
}

func detectMimeType(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer), nil
}

func generateId() string {
	buffer := make([]byte, 4)
	rand.Read(buffer)

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buffer)
}

func createImage(config *Config, body io.ReadCloser, fileHeader *multipart.FileHeader) (Image, error) {
	id := generateId()
	path := filepath.Join(config.SourceFolder, id)

	err := writeBody(body, path)
	if err != nil {
		return Image{}, err
	}

	mimeType, err := detectMimeType(path)
	if err != nil {
		return Image{}, err
	}

	image := Image{
		Id:           id,
		OriginalName: filepath.Base(fileHeader.Filename),
		CreatedAt:    time.Now(),
		MimeType:     mimeType,
	}
	return image, nil
}

type UserInfo struct {
	Id    string
	Name  string
	Email string
}

func parseUser(r *http.Request) UserInfo {
	return UserInfo{
		r.Header.Get("X-Auth-Subject"),
		r.Header.Get("X-Auth-Username"),
		r.Header.Get("X-Auth-Email"),
	}
}

type UploadData struct {
	User    UserInfo
	Results []Result
}

func returnResult(w http.ResponseWriter, templateName string, data interface{}) error {
	pageTemplate, err := template.ParseFiles(
		"templates/_base.html",
		"templates/_header.html",
		"templates/_navigation.html",
		"templates/_footer.html",
		fmt.Sprintf("templates/%s", templateName),
	)
	if err != nil {
		return err
	}

	err = pageTemplate.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	config := NewConfigFromEnv()

	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Address,
		Password: config.Redis.Password,
	})

	db, err := sql.Open(config.Database.Format, config.Database.Url)
	if err != nil {
		panic(err)
	}

	imageServer := http.FileServer(http.Dir(config.TargetFolder))
	assetServer := http.FileServer(http.Dir("assets"))

	http.HandleFunc("/upload/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			user := parseUser(r)

			err := r.ParseMultipartForm(32 << 20)
			if err != nil {
				if err = returnResult(w, "upload.html", UploadData{
					user,
					[]Result{{
						Success: false,
						Errors:  []string{err.Error()},
					}},
				}); err != nil {
					panic(err)
				}
			}

			var images []Image
			var ids []string

			m := r.MultipartForm
			files := m.File["file"]
			for _, header := range files {
				file, err := header.Open()
				if err != nil {
					if err = returnResult(w, "upload.html", UploadData{
						user,
						[]Result{{
							Success: false,
							Errors:  []string{err.Error()},
						}},
					}); err != nil {
						panic(err)
					}
					return
				}
				image, err := createImage(&config, file, header)
				if err != nil {
					if err = returnResult(w, "upload.html", UploadData{
						user,
						[]Result{{
							Success: false,
							Errors:  []string{err.Error()},
						}},
					}); err != nil {
						panic(err)
					}
					return
				}

				images = append(images, image)
				ids = append(ids, image.Id)
			}

			pubsub := client.Subscribe(config.ResultChannel)
			waiting := make(map[string]bool)
			for _, image := range images {
				_, err = db.Exec("INSERT INTO images (id, owner, created_at, original_name, type) VALUES ($1, $2, $3, $4, $5)", image.Id, user.Id, image.CreatedAt, image.OriginalName, image.MimeType)
				if err != nil {
					panic(err)
				}

				data, err := json.Marshal(image)
				if err != nil {
					if err = returnResult(w, "upload.html", UploadData{
						user,
						[]Result{{
							Success: false,
							Errors:  []string{err.Error()},
						}},
					}); err != nil {
						panic(err)
					}
					return
				}

				fmt.Printf("Created task %s at %d\n", image.Id, time.Now().Unix())
				client.RPush(fmt.Sprintf("queue:%s", config.ImageQueue), data)
				fmt.Printf("Submitted task %s at %d\n", image.Id, time.Now().Unix())

				waiting[image.Id] = true
			}

			var results []Result
			for len(waiting) != 0 {
				message, err := pubsub.ReceiveMessage()
				if err != nil {
					if err = returnResult(w, "upload.html", UploadData{
						user,
						[]Result{{
							Success: false,
							Errors:  []string{err.Error()},
						}},
					}); err != nil {
						panic(err)
					}
					return
				}

				result := Result{}
				err = json.Unmarshal([]byte(message.Payload), &result)
				if err != nil {
					if err = returnResult(w, "upload.html", UploadData{
						user,
						[]Result{{
							Success: false,
							Errors:  []string{err.Error()},
						}},
					}); err != nil {
						panic(err)
					}
					return
				}

				fmt.Printf("Returned task %s at %d\n", result.Id, time.Now().Unix())

				if _, ok := waiting[result.Id]; ok {
					delete(waiting, result.Id)

					results = append(results, result)
				}
			}

			if err = returnResult(w, "upload.html", UploadData{
				user,
				results,
			}); err != nil {
				panic(err)
			}
			return
		} else {
			user := parseUser(r)
			if err = returnResult(w, "upload.html", UploadData{
				user,
				[]Result{},
			}); err != nil {
				panic(err)
			}
		}
	})

	http.HandleFunc("/me/images/", func(w http.ResponseWriter, r *http.Request) {
		user := parseUser(r)

		type ImageListData struct {
			User   UserInfo
			Images []Image
		}

		result, err := db.Query(`
			SELECT
				id,
				coalesce(title,  ''),
				coalesce(description, ''),
        		coalesce(created_at, to_timestamp(0)),
				coalesce(original_name, ''),
				coalesce(type, '')
			FROM images
			WHERE owner = $1
			`, user.Id)
		if err != nil {
			panic(err)
		}

		var images []Image
		for result.Next() {
			var info Image
			err := result.Scan(&info.Id, &info.Title, &info.Description, &info.CreatedAt, &info.OriginalName, &info.MimeType)
			if err != nil {
				panic(err)
			}
			images = append(images, info)
		}

		if err = returnResult(w, "me_images.html", ImageListData{
			user,
			images,
		}); err != nil {
			panic(err)
		}
	})

	http.Handle("/assets/", http.StripPrefix("/assets/", assetServer))
	http.Handle("/i/", http.StripPrefix("/i/", imageServer))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		user := parseUser(r)

		type IndexData struct {
			User UserInfo
		}

		if err = returnResult(w, "index.html", IndexData{
			user,
		}); err != nil {
			panic(err)
		}
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
