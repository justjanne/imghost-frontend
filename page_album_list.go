package main

import (
	"net/http"
)

func pageAlbumList(ctx PageContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}
