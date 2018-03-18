package main

import (
	"net/http"
)

type ImageListData struct {
	User   UserInfo
	Images []Image
}

func pageImageList(ctx PageContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := parseUser(r)

		result, err := ctx.Database.Query(`
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

		if err = formatTemplate(w, "image_list.html", ImageListData{
			user,
			images,
		}); err != nil {
			panic(err)
		}
	})
}
