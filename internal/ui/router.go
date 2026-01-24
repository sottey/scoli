package ui

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

//go:embed web/*
var assets embed.FS

func NewRouter() chi.Router {
	r := chi.NewRouter()

	fsys, err := fs.Sub(assets, "web")
	if err != nil {
		panic(err)
	}

	auth := loadAuthConfig()
	if auth.enabled {
		r.Use(auth.middleware)
	}

	fileServer := http.FileServer(http.FS(fsys))

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		if !auth.enabled {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		if auth.validSession(r) {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		serveLoginPage(w, r, fsys)
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		if !auth.enabled {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if auth.enabled && checkPassword(auth.password, r.FormValue("password")) {
			if err := auth.issueSessionCookie(w, r); err != nil {
				http.Error(w, "unable to create session", http.StatusInternalServerError)
				return
			}
			next := sanitizeNextPath(r.FormValue("next"))
			http.Redirect(w, r, next, http.StatusFound)
			return
		}
		next := sanitizeNextPath(r.FormValue("next"))
		target := "/login?error=1"
		if next != "/" && next != "" {
			target = target + "&next=" + url.QueryEscape(next)
		}
		http.Redirect(w, r, target, http.StatusFound)
	})

	r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		if !auth.enabled {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		auth.clearSessionCookie(w, r)
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(fsys, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(data))
	})

	r.Get("/icons/*", func(w http.ResponseWriter, r *http.Request) {
		iconPath, ok := sanitizeIconPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if serveIconFromDisk(w, r, iconPath) {
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	r.Handle("/*", fileServer)

	return r
}

func sanitizeIconPath(requestPath string) (string, bool) {
	path := strings.TrimPrefix(requestPath, "/icons/")
	if path == "" {
		return "", false
	}
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", false
		}
	}
	return filepath.Clean(path), true
}

func serveIconFromDisk(w http.ResponseWriter, r *http.Request, iconPath string) bool {
	iconsDir := filepath.Join("internal", "ui", "web", "icons")
	target := filepath.Join(iconsDir, iconPath)
	cleanDir := filepath.Clean(iconsDir) + string(filepath.Separator)
	if !strings.HasPrefix(filepath.Clean(target), cleanDir) {
		return false
	}
	info, err := os.Stat(target)
	if err != nil || info.IsDir() {
		return false
	}
	http.ServeFile(w, r, target)
	return true
}

func serveLoginPage(w http.ResponseWriter, r *http.Request, fsys fs.FS) {
	data, err := fs.ReadFile(fsys, "login.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, "login.html", time.Now(), bytes.NewReader(data))
}
