package main

import (
	"net/http"
	"fmt"
	"path"
)

type AlbumDetailData struct {
	User   UserInfo
	Album  Album
	IsMine bool
}

func pageAlbumDetail(ctx PageContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := parseUser(r)

		_, albumId := path.Split(r.URL.Path)
		result, err := ctx.Database.Query(`
			SELECT
				id,
				owner,
				coalesce(title,  ''),
				coalesce(description, ''),
        		coalesce(created_at, to_timestamp(0))
			FROM albums
			WHERE id = $1
			`, albumId)
		if err != nil {
			panic(err)
		}

		var info Album
		if result.Next() {
			var owner string
			err := result.Scan(&info.Id, &owner, &info.Title, &info.Description, &info.CreatedAt)
			if err != nil {
				panic(err)
			}

			result, err := ctx.Database.Query(`
			SELECT
				image,
				title,
				description,
				position
			FROM album_images
			WHERE album = $1
			ORDER BY position ASC
			`, albumId)
			if err != nil {
				panic(err)
			}

			for result.Next() {
				var image AlbumImage
				err := result.Scan(&image.Id, &owner, &image.Title, &image.Description, &image.Position)
				if err != nil {
					panic(err)
				}

				info.Images = append(info.Images, image)
			}

			if err = formatTemplate(w, "album_detail.html", AlbumDetailData{
				user,
				info,
				owner == user.Id,
			}); err != nil {
				panic(err)
			}

			return
		}

		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Album not found")
	})
}
