package main

import (
	"net/http"
	"strings"
)

type IndexData struct {
	User UserInfo
}

func removeFileExtensions(path string) string {
	var i = strings.IndexByte(path, '.')
	if i < 0 {
		return path
	} else {
		return path[0:i]
	}
}

func pageIndex(ctx PageContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/me/images", http.StatusFound)
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
			r.URL.Path = removeFileExtensions(r.URL.Path)
			ctx.Images.ServeHTTP(w, r)
		}
	})
}
