package main

import (
	"fmt"
	"net/http"
	"html/template"
	"time"
	"database/sql"
	"github.com/go-redis/redis"
	"strings"
	"encoding/json"
)

type UserInfo struct {
	Id    string
	Name  string
	Email string
	Roles []string
}

func (info UserInfo) HasRole(role string) bool {
	for _, r := range info.Roles {
		if r == role {
			return true
		}
	}
	return false
}

type PageContext struct {
	Config      *Config
	Redis       *redis.Client
	Database    *sql.DB
	Images      http.Handler
	AssetServer http.Handler
}

type AlbumImage struct {
	Id          string
	Title       string
	Description string
	Position    int
}

type Album struct {
	Id          string
	Title       string
	Description string
	CreatedAt   time.Time
	Images      []AlbumImage
}

func parseUser(r *http.Request) UserInfo {
	return UserInfo{
		r.Header.Get("X-Auth-Subject"),
		r.Header.Get("X-Auth-Username"),
		r.Header.Get("X-Auth-Email"),
		strings.Split(r.Header.Get("X-Auth-Roles"), ","),
	}
}

func returnJson(w http.ResponseWriter, data interface{}) error {
	marshalled, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(marshalled); err != nil {
		return err
	}

	return nil
}

func formatTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	pageTemplate, err := template.ParseFiles(
		"templates/_base.html",
		"templates/_header.html",
		"templates/_navigation.html",
		"templates/_footer.html",
		fmt.Sprintf("templates/%s", templateName),
	)
	if err != nil {
		return err
	}

	err = pageTemplate.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}
