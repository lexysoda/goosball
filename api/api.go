package api

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/lexysoda/goosball/controller"
)

type Api struct {
	*http.ServeMux
	controller *controller.Controller
}

func New(c *controller.Controller) *Api {
	a := &Api{}
	mux := http.NewServeMux()
	mux.HandleFunc("/state", a.GetState)
	mux.HandleFunc("/score", a.Score)
	mux.HandleFunc("/queue", a.Queue)
	mux.HandleFunc("/queue/", a.RemoveFromQueue)
	mux.HandleFunc("/users", a.Users)
	mux.HandleFunc("/set", a.CancelSet)
	a.ServeMux = mux
	a.controller = c
	return a
}

func (a *Api) GetState(w http.ResponseWriter, r *http.Request) {
	json, err := json.Marshal(a.controller.State)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

func (a *Api) Score(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only posts", http.StatusMethodNotAllowed)
		return
	}
	var isTeamA bool
	if err := json.NewDecoder(r.Body).Decode(&isTeamA); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.controller.Score(isTeamA)
}

func (a *Api) Queue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only posts", http.StatusMethodNotAllowed)
		return
	}
	var userId string
	if err := json.NewDecoder(r.Body).Decode(&userId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.controller.AddToQueue(userId)
}

func (a *Api) Users(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only get", http.StatusBadRequest)
		return
	}
	users, err := a.controller.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

func (a *Api) RemoveFromQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "only delete", http.StatusBadRequest)
		return
	}
	id := path.Base(r.URL.Path)
	if id == "queue" {
		http.Error(w, "need id", http.StatusBadRequest)
		return
	}
	a.controller.RemoveFromQueue(id)
	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) CancelSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "only delete", http.StatusBadRequest)
		return
	}
	a.controller.CancelSet()
	w.WriteHeader(http.StatusNoContent)
}
