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
	User   UserInfo
	Result Result
}

func returnResult(w http.ResponseWriter, data UploadData) error {
	var pageTemplate *template.Template
	var err error
	if data.Result.Success {
		pageTemplate, err = template.New("upload_success.html").ParseFiles("templates/upload_success.html")
	} else {
		pageTemplate, err = template.New("upload_failure.html").ParseFiles("templates/upload_failure.html")
	}
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

			r.ParseMultipartForm(32 << 20)
			file, header, err := r.FormFile("file")
			image, err := createImage(&config, file, header)
			if err != nil {
				returnResult(w, UploadData{
					user,
					Result{
						Id:      "",
						Success: false,
						Errors:  []string{err.Error()},
					},
				})
				return
			}

			_, err = db.Exec("INSERT INTO images (id, owner, created_at, original_name, type) VALUES ($1, $2, $3, $4, $5)", image.Id, user.Id, image.CreatedAt, image.OriginalName, image.MimeType)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Created task %s at %d\n", image.Id, time.Now().Unix())

			data, err := json.Marshal(image)
			if err != nil {
				returnResult(w, UploadData{
					user,
					Result{
						Id:      image.Id,
						Success: false,
						Errors:  []string{err.Error()},
					},
				})
				return
			}

			pubsub := client.Subscribe(config.ResultChannel)
			client.RPush(fmt.Sprintf("queue:%s", config.ImageQueue), data)

			fmt.Printf("Submitted task %s at %d\n", image.Id, time.Now().Unix())

			waiting := true
			for waiting {
				message, err := pubsub.ReceiveMessage()
				if err != nil {
					returnResult(w, UploadData{
						user,
						Result{
							Id:      image.Id,
							Success: false,
							Errors:  []string{err.Error()},
						},
					})
					return
				}

				result := Result{}
				err = json.Unmarshal([]byte(message.Payload), &result)
				if err != nil {
					returnResult(w, UploadData{
						user,
						Result{
							Id:      image.Id,
							Success: false,
							Errors:  []string{err.Error()},
						},
					})
					return
				}

				fmt.Printf("Returned task %s at %d\n", result.Id, time.Now().Unix())

				if result.Id == image.Id {
					waiting = false

					returnResult(w, UploadData{
						user,
						result,
					})
					return
				}
			}
		} else {
			user := parseUser(r)

			type UploadData struct {
				User UserInfo
			}

			tmpl, err := template.New("upload.html").ParseFiles("templates/upload.html")
			if err != nil {
				panic(err)
			}
			err = tmpl.Execute(w, UploadData{
				user,
			})
			if err != nil {
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

		pageTemplate, err := template.New("me_images.html").ParseFiles("templates/me_images.html")
		if err != nil {
			panic(err)
		}
		err = pageTemplate.Execute(w, ImageListData{
			user,
			images,
		})
		if err != nil {
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

		tmpl, err := template.New("index.html").ParseFiles("templates/index.html")
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(w, IndexData{
			user,
		})
		if err != nil {
			panic(err)
		}
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
