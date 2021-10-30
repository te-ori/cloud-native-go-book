package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func helloMuxHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Write([]byte("Hello gorilla/mux!" + vars["some"]))
}

func handleByGorilla() {
	r := mux.NewRouter()
	r.HandleFunc("/{some}", helloMuxHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
