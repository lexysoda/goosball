package main

import (
	"log"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/lexysoda/goskill"

	"github.com/lexysoda/goosball/api"
	"github.com/lexysoda/goosball/controller"
	"github.com/lexysoda/goosball/db"
	sapi "github.com/lexysoda/goosball/slack/api"
	"github.com/lexysoda/goosball/slack/bot"
)

type config struct {
	SlackHomeChannel string `required:"true" split_words:"true"`
	Address          string `default:":8080"`
	DBPath           string `default:"./goosball.db" split_words:"true"`
	SlackToken       string `required:"true" split_words:"true"`
	SlackAppToken    string `required:"true" split_words:"true"`
}

func main() {
	config := config{}
	if err := envconfig.Process("goosball", &config); err != nil {
		log.Fatal(err)
	}
	log.Println(config.Address)
	db, err := db.New(config.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	c := &controller.Controller{
		Db:        db,
		Elo:       goskill.New(),
		SlackAPI:  sapi.New(config.SlackToken),
		SlackHome: config.SlackHomeChannel,
	}
	_ = bot.New(c, config.SlackToken, config.SlackAppToken)
	a := api.New(c)
	http.Handle("/api/", http.StripPrefix("/api", a))
	http.Handle("/", http.FileServer(http.Dir("static")))
	log.Fatal(http.ListenAndServe(config.Address, nil))
}
