package routers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
)

func GitWebhooks(r chi.Router) {
	r.Post("/", postWebhook)
}

func postWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "No request body", 400)
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Unable to parse request body", 500)
	}

	var webhookData WebhookData
	err = json.Unmarshal(body, &webhookData)

	if err != nil {
		http.Error(w, "Unable to process git data", 500)
	}

	webhookData.Configure()
	err = webhookData.Repository.UpdateOrClone()

	if err != nil {
		http.Error(w, "Unable to update or clone repository", 500)
	}

	w.WriteHeader(200)
}
