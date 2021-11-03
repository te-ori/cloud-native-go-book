package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var throttled = Throttle(getHostName, 1, 1, time.Second)

func getHostName(ctx context.Context) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	return os.Hostname()
}

func throttledHandler(w http.ResponseWriter, r *http.Request) {
	ok, hostname, err := throttled(r.Context(), r.RemoteAddr)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !ok {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(hostname))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/hostname", throttledHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
