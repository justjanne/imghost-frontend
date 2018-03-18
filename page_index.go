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
		} else {
			ctx.Images.ServeHTTP(w, r)
		}
	})
}
