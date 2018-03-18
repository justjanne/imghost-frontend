package main

import (
	"database/sql"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"net/http"
	"fmt"
)

func headerWrapper(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println(r.URL.Path)
		for key, value := range r.Header {
			fmt.Printf("%s: %s", key, value)
		}
		handler.ServeHTTP(w, r)
	})
}

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

	http.Handle("/upload/", headerWrapper(pageUpload(pageContext)))

	http.Handle("/i/", headerWrapper(http.StripPrefix("/i/", pageImageDetail(pageContext))))
	http.Handle("/me/i/", headerWrapper(http.StripPrefix("/me/i/", pageImageDetail(pageContext))))

	http.Handle("/a/", headerWrapper(http.StripPrefix("/a/", pageAlbumDetail(pageContext))))
	http.Handle("/me/a/", headerWrapper(http.StripPrefix("/me/a/", pageAlbumDetail(pageContext))))

	http.Handle("/me/images/", headerWrapper(pageImageList(pageContext)))
	http.Handle("/assets/", headerWrapper(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))))
	http.Handle("/", headerWrapper(pageIndex(pageContext)))

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
