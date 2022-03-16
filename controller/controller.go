package controller

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/lexysoda/goosball/db"
	"github.com/lexysoda/goosball/model"
	"github.com/lexysoda/goosball/slack/api"
	"github.com/lexysoda/goskill"
)

type Controller struct {
	Db       db.DB
	State    State
	Elo      goskill.BTFull
	SlackAPI *api.Slack
	sync.Mutex
	SlackHome string
}

type State struct {
	Queue []model.User
	Games []model.Game
	Set   *model.Set
}

func (c *Controller) GetAllUsers() ([]model.User, error) {
	return c.Db.GetUsers()
}

func (c *Controller) getOrCreateUser(id string) (*model.User, error) {
	uGot, err := c.Db.GetUser(id)
	if err == nil {
		return uGot, nil
	} else if !errors.Is(err, db.NoRow) {
		return nil, err
	}
	uNew, err := c.SlackAPI.GetUser(id)
	if err != nil {
		return nil, err
	}
	uNew.Goskill = c.Elo.Skill()
	return uNew, c.Db.AddUser(uNew)
}

func (c *Controller) AddToQueue(ids ...string) {
	c.Lock()
	defer c.Unlock()
loop:
	for _, id := range ids {
		user, err := c.getOrCreateUser(id)
		if err != nil {
			log.Println("Failed to get or create user %s: %s\n", id, err)
		}
		for _, uq := range c.State.Queue {
			if uq.ID == user.ID {
				continue loop
			}
		}
		c.State.Queue = append(c.State.Queue, *user)
	}
	if len(c.State.Queue) >= 4 && c.State.Set == nil {
		c.StartMatch()
		return
	}
	c.SendQueueSlack()
	return
}

func (c *Controller) RemoveFromQueue(id string) {
	c.Lock()
	defer c.Unlock()
	newQueue := []model.User{}
	for _, u := range c.State.Queue {
		if u.ID != id {
			newQueue = append(newQueue, u)
		}
	}
	c.State.Queue = newQueue
	c.SendQueueSlack()
}

func (c *Controller) StartMatch() {
	p := c.State.Queue[:4]
	c.State.Queue = c.State.Queue[4:]
	m1 := math.Abs(c.Elo.WinProbability([]goskill.Skiller{p[0], p[1]}, []goskill.Skiller{p[2], p[3]}) - 0.5)
	m2 := math.Abs(c.Elo.WinProbability([]goskill.Skiller{p[0], p[2]}, []goskill.Skiller{p[1], p[3]}) - 0.5)
	m3 := math.Abs(c.Elo.WinProbability([]goskill.Skiller{p[0], p[3]}, []goskill.Skiller{p[1], p[2]}) - 0.5)

	perm := [4]int{0, 3, 1, 2}
	if m1 <= m2 && m1 <= m3 {
		perm = [4]int{0, 1, 2, 3}
	} else if m2 <= m1 && m2 <= m3 {
		perm = [4]int{0, 2, 1, 3}
	}
	Set := model.Set{
		P1: p[perm[0]],
		P2: p[perm[1]],
		P3: p[perm[2]],
		P4: p[perm[3]],
	}
	c.State.Set = &Set
	log.Printf("%s and %s playing against %s and %s\n",
		Set.P1.DisplayName,
		Set.P2.DisplayName,
		Set.P3.DisplayName,
		Set.P4.DisplayName,
	)
	c.SlackAPI.Send(c.SlackHome,
		fmt.Sprintf("A new game started: <@%s> <@%s> vs <@%s> <@%s>",
			Set.P1.ID,
			Set.P2.ID,
			Set.P3.ID,
			Set.P4.ID,
		),
	)
	c.NewGame()
}

func (c *Controller) NewGame() {
	if c.State.Set == nil {
		log.Println("Trying to start game without Set.")
		return
	}
	c.State.Games = append(c.State.Games, model.Game{
		StartTime: time.Now(),
	})
}

func (c *Controller) Score(isTeamA bool) error {
	c.Lock()
	defer c.Unlock()
	if len(c.State.Games) == 0 {
		return fmt.Errorf("Tried to score but no game is running")
	}
	g := &c.State.Games[len(c.State.Games)-1]
	if isTeamA {
		g.GoalsA += 1
	} else {
		g.GoalsB += 1
	}
	if g.GoalsA == 6 || g.GoalsB == 6 {
		c.FinishGame(isTeamA)
	}
	return nil
}

func (c *Controller) FinishGame(isTeamA bool) error {
	g := &c.State.Games[len(c.State.Games)-1]
	g.EndTime = time.Now()
	if g.GoalsA == 0 || g.GoalsB == 0 {
		p1, p2 := c.State.Set.P1, c.State.Set.P2
		if isTeamA {
			p1, p2 = c.State.Set.P3, c.State.Set.P4
		}
		c.SlackAPI.Send(c.SlackHome,
			fmt.Sprintf(
				"<@%s> and <@%s> have to crawl. shame.",
				p1.ID,
				p2.ID,
			),
		)
	}
	if len(c.State.Games) == 2 &&
		(isTeamA && c.State.Games[0].GoalsA == 6 ||
			!isTeamA && c.State.Games[0].GoalsB == 6) {
		return c.FinishSet(isTeamA)
	} else if len(c.State.Games) == 3 {
		return c.FinishSet(isTeamA)
	}

	c.NewGame()
	return nil
}

func (c *Controller) FinishSet(isTeamA bool) error {
	SetID, err := c.Db.NewSet(c.State.Set)
	if err != nil {
		log.Fatal(err)
	}
	for _, g := range c.State.Games {
		g.SetID = SetID
		_, err = c.Db.NewGame(&g)
		if err != nil {
			log.Fatal(err)
		}
	}
	teamA := []goskill.Skiller{c.State.Set.P1, c.State.Set.P2}
	teamB := []goskill.Skiller{c.State.Set.P3, c.State.Set.P4}
	if isTeamA {
		c.SlackAPI.Send(c.SlackHome,
			fmt.Sprintf(
				"<@%s> and <@%s> won against <@%s> and <@%s>!",
				c.State.Set.P1.ID,
				c.State.Set.P2.ID,
				c.State.Set.P3.ID,
				c.State.Set.P4.ID,
			),
		)
		c.Elo.Rank([][]goskill.Skiller{teamA, teamB})
	} else {
		c.SlackAPI.Send(c.SlackHome,
			fmt.Sprintf(
				"<@%s> and <@%s> won against <@%s> and <@%s>!",
				c.State.Set.P3.ID,
				c.State.Set.P4.ID,
				c.State.Set.P1.ID,
				c.State.Set.P2.ID,
			),
		)
		c.Elo.Rank([][]goskill.Skiller{teamB, teamA})
	}
	c.Db.UpdateUser(&c.State.Set.P1)
	c.Db.UpdateUser(&c.State.Set.P2)
	c.Db.UpdateUser(&c.State.Set.P3)
	c.Db.UpdateUser(&c.State.Set.P4)

	c.State.Games = []model.Game{}
	c.State.Set = nil

	return nil
}

func (c *Controller) CancelSet() {
	c.Lock()
	c.State.Set = nil
	c.State.Games = []model.Game{}
	c.Unlock()
	c.SlackAPI.Send(c.SlackHome, "The current game was canceled.")
	c.SendQueueSlack()
}

func (c *Controller) SendQueueSlack() {
	if len(c.State.Queue) == 0 {
		err := c.SlackAPI.Send(c.SlackHome, "The queue is empty.")
		if err != nil {
			log.Printf("Failed to send slack message: %s\n", err)
		}
		return
	}
	message := "Current queue: "
	for i, u := range c.State.Queue {
		message += "<@" + u.ID + ">"
		if i != len(c.State.Queue)-1 {
			message += ", "
		}
	}
	message += "."
	err := c.SlackAPI.Send(c.SlackHome, message)
	if err != nil {
		log.Printf("Failed to send slack message: %s\n", err)
	}
}
