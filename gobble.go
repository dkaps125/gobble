package main

import (
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "No body", 400)
		}

		body, _ := ioutil.ReadAll(r.Body)

		w.Write(body)
	})

	http.ListenAndServe(":3000", r)
}
