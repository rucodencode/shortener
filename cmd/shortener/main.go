package main

import (
	"net/http"

	"github.com/rucodencode/shortener/internal/handler"
	"github.com/rucodencode/shortener/internal/repository"
	"github.com/rucodencode/shortener/internal/service"
)

func main() {
	addr := ":8080"
	baseURL := "http://localhost:8080"

	repo := repository.NewLinkRepository()
	linkService := service.NewLinkService(repo)
	linkHandler := handler.NewLinkHandler(linkService, baseURL)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			if r.Method != http.MethodPost {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}

			linkHandler.Create(w, r)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		linkHandler.Get(w, r)
	})

	err := http.ListenAndServe(addr, mux)

	if err != nil {
		panic(err)
	}
}
