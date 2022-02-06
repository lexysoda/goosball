package api

import (
	"encoding/json"
	"net/http"

	"github.com/lexysoda/goosball/controller"
)

type Api struct {
	*http.ServeMux
	controller *controller.Controller
}

func (a *Api) Init(c *controller.Controller) {
	mux := http.NewServeMux()
	mux.HandleFunc("/state", a.GetState)
	mux.HandleFunc("/score", a.Score)
	a.ServeMux = mux
	a.controller = c
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
