package main

import (
	"net/http"
)

type IndexData struct {
	User UserInfo
}

func pageIndex(ctx PageContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			user := parseUser(r)

			if err := formatTemplate(w, "index.html", IndexData{
				user,
			}); err != nil {
				panic(err)
			}
        } else if r.URL.Path == "/favicon.ico" {
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			ctx.AssetServer.ServeHTTP(w, r)
        } else if r.URL.Path == "/favicon.png" {
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			ctx.AssetServer.ServeHTTP(w, r)
        } else if r.URL.Path == "/favicon.svg" {
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			ctx.AssetServer.ServeHTTP(w, r)
		} else {
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			ctx.Images.ServeHTTP(w, r)
		}
	})
}
