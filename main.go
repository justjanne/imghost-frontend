package main

import (
	"os"
	"net/http"
	"io"
	"path/filepath"
	"encoding/json"
	"github.com/go-redis/redis"
	"fmt"
	"encoding/base64"
	"crypto/rand"
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

func createImage(config *Config, body io.ReadCloser) (Image, error) {
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
		Id:       id,
		MimeType: mimeType,
	}
	return image, nil
}

func returnResult(writer http.ResponseWriter, result Result) error {
	writer.Header().Add("Content-Type", "text/html")
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}
	writer.Write([]byte("<pre>"))
	writer.Write(body)
	writer.Write([]byte("</pre>"))
	if result.Success {
		writer.Write([]byte("<p><a href=\""))
		writer.Write([]byte(fmt.Sprintf("https://i.k8r.eu/i/%s", result.Id)))
		writer.Write([]byte("\">Uploaded Image</a></p>"))
	}
	return nil
}

func printHeaders(r *http.Request) {
	fmt.Println(r.URL)
	for key, value := range r.Header {
		fmt.Printf("  %s: %s\n", key, value)
	}
}

func main() {
	config := NewConfigFromEnv()

	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Address,
		Password: config.Redis.Password,
	})

	//db, err := sql.Open(config.Database.Format, config.Database.Url)
	//if err != nil {
	//panic(err)
	//}

	staticServer := http.FileServer(http.Dir("static/"))
	imageServer := http.FileServer(http.Dir(config.TargetFolder))

	http.HandleFunc("/upload/", func(w http.ResponseWriter, r *http.Request) {
		printHeaders(r)

		if r.Method == "POST" {
			r.ParseMultipartForm(32 << 20)
			file, _, err := r.FormFile("file")
			image, err := createImage(&config, file)
			if err != nil {
				returnResult(w, Result{
					Id:      "",
					Success: false,
					Errors:  []string{err.Error()},
				})
				return
			}

			fmt.Printf("Created task %s at %d\n", image.Id, time.Now().Unix())

			data, err := json.Marshal(image)
			if err != nil {
				returnResult(w, Result{
					Id:      image.Id,
					Success: false,
					Errors:  []string{err.Error()},
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
					returnResult(w, Result{
						Id:      image.Id,
						Success: false,
						Errors:  []string{err.Error()},
					})
					return
				}

				result := Result{}
				err = json.Unmarshal([]byte(message.Payload), &result)
				if err != nil {
					returnResult(w, Result{
						Id:      image.Id,
						Success: false,
						Errors:  []string{err.Error()},
					})
					return
				}

				fmt.Printf("Returned task %s at %d\n", result.Id, time.Now().Unix())

				if result.Id == image.Id {
					waiting = false

					returnResult(w, result)
					return
				}
			}
		} else {
			staticServer.ServeHTTP(w, r)
		}
	})

	http.Handle("/i/", http.StripPrefix("/i/", imageServer))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		printHeaders(r)
		staticServer.ServeHTTP(w, r)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
