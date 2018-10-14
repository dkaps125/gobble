package webhooks

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"gobble/utils"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi"
)

type GitWebhook struct {
	Repository Repo `json:"repository"`
	secret     []byte
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

	if utils.HTTPErrorCheck(err, w, 500) {
		return
	}

	var webhook GitWebhook

	auth, err := webhook.checkSecret(r.Header, body)

	if !auth || err != nil {
		log.Println(err)
		http.Error(w, "Repository authentication failed", 500)
		return
	}

	err = json.Unmarshal(body, &webhook)

	if utils.HTTPErrorCheck(err, w, 500) {
		return
	}

	webhook.Configure()
	err = webhook.Repository.UpdateOrClone()

	if utils.HTTPErrorCheck(err, w, 500) {
		return
	}

	err = webhook.Repository.ImportConfig()

	if err == nil {
		err = webhook.Repository.Deploy()

		if utils.HTTPErrorCheck(err, w, 500) {
			return
		}
	}

	w.WriteHeader(200)
}

func (w *GitWebhook) checkSecret(header http.Header, requestBody []byte) (bool, error) {
	macStr := header.Get("X-Hub-Signature")

	if macStr != "" {
		macStr = strings.Split(macStr, "=")[1]

		messageMac, err := hex.DecodeString(macStr)
		if err != nil {
			return false, err
		}

		secret := w.secret

		if len(secret) == 0 {
			secret = utils.Config.Secret

			if len(secret) == 0 {
				return false, utils.ERRNOCONFIG
			}
		}

		mac := hmac.New(sha1.New, secret)
		mac.Write(requestBody)
		expectedMac := mac.Sum(nil)
		return hmac.Equal(messageMac, expectedMac), nil
	}

	if len(utils.Config.Secret) > 0 {
		return false, utils.ERRGITWEBHOOK{
			GitAction: utils.GITHOOK,
			Message:   "No signature provided",
		}
	} else {
		// this indicates that no secret is required, and none was provided
		return true, nil
	}
}
