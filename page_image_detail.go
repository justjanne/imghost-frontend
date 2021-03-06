package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
	"os"
	"path"
)

type ImageDetailData struct {
	User   UserInfo
	Image  Image
	IsMine bool
}

func pageImageDetail(ctx PageContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := parseUser(r)
		_, imageId := path.Split(r.URL.Path)

		result, err := ctx.Database.Query(`
			SELECT
				id,
				owner,
				coalesce(title,  ''),
				coalesce(description, ''),
        		coalesce(created_at, to_timestamp(0)),
				coalesce(original_name, ''),
				coalesce(type, '')
			FROM images
			WHERE id = $1
			`, imageId)
		if err != nil {
			fmt.Printf("An error occured: %s", err.Error())
			_ = returnError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		var info Image

		if result.Next() {
			var owner string
			err := result.Scan(&info.Id, &owner, &info.Title, &info.Description, &info.CreatedAt, &info.OriginalName, &info.MimeType)
			if err != nil {
				fmt.Printf("An error occured: %s", err.Error())
				_ = returnError(w, http.StatusInternalServerError, "Internal Server Error")
				return
			}

			switch r.PostFormValue("action") {
			case "update":
				_, err = ctx.Database.Exec(
					"UPDATE images SET title = $1, description = $2 WHERE id = $3 AND owner = $4",
					r.PostFormValue("title"),
					r.PostFormValue("description"),
					info.Id,
					user.Id,
				)
				if err != nil {
					fmt.Printf("An error occured: %s", err.Error())
					_ = returnError(w, http.StatusInternalServerError, "Internal Server Error")
					return
				}
				if r.PostFormValue("from_js") == "true" {
					_ = returnJson(w, true)
				} else {
					http.Redirect(w, r, r.URL.Path, http.StatusFound)
				}
				return
			case "delete":
				_, err = ctx.Database.Exec("DELETE FROM images WHERE id = $1 AND owner = $2", info.Id, user.Id)
				if err != nil {
					fmt.Printf("An error occured: %s", err.Error())
					_ = returnError(w, http.StatusInternalServerError, "Internal Server Error")
					return
				}
				for _, definition := range ctx.Config.Sizes {
					err := os.Remove(path.Join(ctx.Config.TargetFolder, fmt.Sprintf("%s%s", info.Id, definition.Suffix)))
					if err != nil && !os.IsNotExist(err) {
						fmt.Printf("An error occured: %s", err.Error())
						_ = returnError(w, http.StatusInternalServerError, "Internal Server Error")
						return
					}
				}
				http.Redirect(w, r, "/me/images", http.StatusFound)
				return
			}

			if err = formatTemplate(w, "image_detail.html", ImageDetailData{
				user,
				info,
				owner == user.Id,
			}); err != nil {
				panic(err)
			}
			return
		}

		_ = returnError(w, http.StatusNotFound, "Image Not Found")
	})
}
