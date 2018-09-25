package routers

import (
	"gobble/routers/webhooks"

	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Mount("/gitwebhook", webhooks.Routes())

	return r
}
