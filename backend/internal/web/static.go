package web

import (
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Register installs the SPA served from ./dist at the root. API routes must be
// mounted before calling Register (Gin falls back to NoRoute for unmatched
// paths). If dist is missing (e.g. `go run` with no build artifact), this is
// a silent no-op.
func Register(r *gin.Engine) {
	dir := locateDist()
	if dir == "" {
		return
	}
	sub := os.DirFS(dir)
	r.NoRoute(func(c *gin.Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if data, err := fs.ReadFile(sub, p); err == nil {
			ctype := mime.TypeByExtension(filepath.Ext(p))
			if ctype == "" {
				ctype = http.DetectContentType(data)
			}
			c.Data(http.StatusOK, ctype, data)
			return
		}
		data, err := fs.ReadFile(sub, "index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
}

// locateDist prefers ./dist next to the running binary so the exe+dist pair
// stays portable across working directories. Falls back to cwd-relative ./dist
// for `go run` during development (the dev binary lives in a temp dir).
func locateDist() string {
	if exe, err := os.Executable(); err == nil {
		if p := filepath.Join(filepath.Dir(exe), "dist"); isDir(p) {
			return p
		}
	}
	if isDir("dist") {
		return "dist"
	}
	return ""
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
