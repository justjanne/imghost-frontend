package main

import (
	"database/sql"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"net/http"
)

func main() {
	config := NewConfigFromEnv()

	db, err := sql.Open(config.Database.Format, config.Database.Url)
	if err != nil {
		panic(err)
	}

	pageContext := PageContext{
		&config,
		redis.NewClient(&redis.Options{
			Addr:     config.Redis.Address,
			Password: config.Redis.Password,
		}),
		db,
		http.FileServer(http.Dir(config.TargetFolder)),
		http.FileServer(http.Dir("assets")),
	}

	http.Handle("/upload/", pageUpload(pageContext))

	http.Handle("/i/", http.StripPrefix("/i/", pageImageDetail(pageContext)))
	http.Handle("/a/", http.StripPrefix("/a/", pageAlbumDetail(pageContext)))

	http.Handle("/me/images/", pageImageList(pageContext))
	http.Handle("/assets/", http.StripPrefix("/assets/", pageContext.AssetServer))
	http.Handle("/", pageIndex(pageContext))

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
