package webhooks

import (
	"encoding/json"
	"gobble/utils"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/go-chi/chi"
)

type GitWebhook struct {
	Repository Repo `json:"repository"`
}

func (w *GitWebhook) Configure() {
	//TODO: pull config info from DB here as well, if it exists
	w.Repository.SetDirectory(path.Join(utils.Config.GetProjectDir(), w.Repository.Name))
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", postWebhook)

	return r
}

func postWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "No request body", 400)
	}

	body, err := ioutil.ReadAll(r.Body)

	utils.HTTPErrorCheck(err, w, 500)

	var webhook GitWebhook
	err = json.Unmarshal(body, &webhook)

	utils.HTTPErrorCheck(err, w, 500)

	webhook.Configure()
	err = webhook.Repository.UpdateOrClone()

	utils.HTTPErrorCheck(err, w, 500)

	err = webhook.Repository.ImportConfig()

	if err == nil {
		err = webhook.Repository.Deploy()

		utils.HTTPErrorCheck(err, w, 500)
	}

	w.WriteHeader(200)
}
