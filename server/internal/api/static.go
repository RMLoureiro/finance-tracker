package api

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// StaticHandler serve o export estático do Next (produção): ficheiro direto,
// depois caminho+".html" (rotas do App Router) e por fim index.html.
func StaticHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean(r.URL.Path)

		if strings.HasPrefix(p, "/_next/static/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			fs.ServeHTTP(w, r)
			return
		}

		full := filepath.Join(dir, filepath.FromSlash(p))
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		if _, err := os.Stat(full + ".html"); err == nil {
			r.URL.Path = p + ".html"
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}
