package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const MaxQueueDepth = 1000

func CurrentQueueDepth() uint {
	return 10
}

func loadSheddingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CurrentQueueDepth() > MaxQueueDepth {
			log.Println("load shedding engaged")

			http.Error(w, "err.Error()", http.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()
	r.Use(loadSheddingMiddleware)

	log.Fatal(http.ListenAndServe(":8080", r))
}
