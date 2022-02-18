package main

import (
	"log"
	"net/http"
	"os"

	"github.com/lexysoda/goskill"

	"github.com/lexysoda/goosball/api"
	"github.com/lexysoda/goosball/controller"
	"github.com/lexysoda/goosball/db"
	"github.com/lexysoda/goosball/slack"
	sapi "github.com/lexysoda/goosball/slack/api"
)

func main() {
	db, err := db.New()
	if err != nil {
		log.Fatal(err)
	}
	users := []string{"U02QK2J4BRD", "U02NPU059QT", "U029URUKJLF", "UB048064V"}
	s := slack.Init()
	c := &controller.Controller{
		Db:        db,
		Elo:       goskill.New(),
		SlackAPI:  sapi.New(),
		Slack:     s,
		SlackHome: os.Getenv("SLACK_HOME_CHANNEL"),
	}
	c.Init()
	for _, id := range users {
		u, err := c.GetOrCreateUser(id)
		if err != nil {
			log.Fatal(err)
		}
		err = c.AddToQueue(u.ID)
		if err != nil {
			log.Fatal(err)
		}
	}

	for i := 0; i < 10; i++ {
		c.Score(true)
	}

	for _, id := range users {
		uGot, err := c.GetOrCreateUser(id)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("User: %s, Mu: %g, SigSq: %g\n", uGot.RealName, uGot.Skill.Mu, uGot.Skill.SigSq)
	}

	a := &api.Api{}
	a.Init(c)
	http.Handle("/api/", http.StripPrefix("/api", a))
	http.Handle("/", http.FileServer(http.Dir("static")))
	log.Fatal(http.ListenAndServe(":1337", nil))
}
