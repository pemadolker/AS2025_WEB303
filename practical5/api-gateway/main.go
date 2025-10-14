package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Route /api/users/* to user-service
	r.HandleFunc("/api/users*", proxyTo("http://user-service:8081", "/users"))

	// Route /api/menu/* to menu-service
	r.HandleFunc("/api/menu*", proxyTo("http://menu-service:8082", "/menu"))

	// Route /api/orders/* to order-service
	r.HandleFunc("/api/orders*", proxyTo("http://order-service:8083", "/orders"))

	log.Println("API Gateway starting on :8080")
	http.ListenAndServe(":8080", r)
}

func proxyTo(targetURL, stripPrefix string) http.HandlerFunc {
	target, _ := url.Parse(targetURL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {
		// Strip /api prefix
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
		log.Printf("Proxying %s to %s%s", r.Method, targetURL, r.URL.Path)
		proxy.ServeHTTP(w, r)
	}
}
