package main

import (
	"log"
	"net/http"
	"os"

	"github.com/lexysoda/goskill"

	"github.com/lexysoda/goosball/api"
	"github.com/lexysoda/goosball/controller"
	"github.com/lexysoda/goosball/db"
	sapi "github.com/lexysoda/goosball/slack/api"
	"github.com/lexysoda/goosball/slack/bot"
)

func main() {
	db, err := db.New()
	if err != nil {
		log.Fatal(err)
	}
	c := &controller.Controller{
		Db:        db,
		Elo:       goskill.New(),
		SlackAPI:  sapi.New(),
		SlackHome: os.Getenv("SLACK_HOME_CHANNEL"),
	}
	_ = bot.New(c)
	a := api.New(c)
	http.Handle("/api/", http.StripPrefix("/api", a))
	http.Handle("/", http.FileServer(http.Dir("static")))
	log.Fatal(http.ListenAndServe(":1337", nil))
}
