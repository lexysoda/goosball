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
	"github.com/lexysoda/goosball/slack"
	"github.com/lexysoda/goskill"
)

type Controller struct {
	Db    db.DB
	State State
	Elo   goskill.BradleyTerryFull
	Slack *slack.Slack
	sync.Mutex
}

type State struct {
	Queue []model.User
	Games []model.Game
	Set   *model.Set
}

func (c *Controller) GetAllUsers() ([]model.User, error) {
	return c.Db.GetUsers()
}

func (c *Controller) GetOrCreateUser(id string) (*model.User, error) {
	uGot, err := c.Db.GetUser(id)
	if err == nil {
		return uGot, nil
	} else if !errors.Is(err, db.NoRow) {
		return nil, err
	}
	uNew, err := c.Slack.GetUser(id)
	if err != nil {
		return nil, err
	}
	uNew.Skill = c.Elo.Skill()
	return uNew, c.Db.AddUser(uNew)
}

func (c *Controller) AddToQueue(ids ...string) error {
	c.Lock()
	defer c.Unlock()
	for _, id := range ids {
		user, err := c.GetOrCreateUser(id)
		if err != nil {
			return err
		}
		for _, uq := range c.State.Queue {
			if uq.ID == user.ID {
				return fmt.Errorf("User %s already in Queue", uq.DisplayName)
			}
		}
		c.State.Queue = append(c.State.Queue, *user)
	}
	if len(c.State.Queue) >= 4 && c.State.Set == nil {
		c.StartMatch()
	}
	return nil
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
}

func (c *Controller) StartMatch() {
	p := c.State.Queue[:4]
	c.State.Queue = c.State.Queue[4:]
	m1 := math.Abs(c.Elo.WinProbability(
		[]*goskill.Skill{&p[0].Skill, &p[1].Skill}, []*goskill.Skill{&p[2].Skill, &p[3].Skill}) - 0.5)
	m2 := math.Abs(c.Elo.WinProbability(
		[]*goskill.Skill{&p[0].Skill, &p[2].Skill}, []*goskill.Skill{&p[1].Skill, &p[3].Skill}) - 0.5)
	m3 := math.Abs(c.Elo.WinProbability(
		[]*goskill.Skill{&p[0].Skill, &p[3].Skill}, []*goskill.Skill{&p[1].Skill, &p[2].Skill}) - 0.5)

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
	log.Printf("%s and %s playing against %s and %s\n", Set.P1.DisplayName, Set.P2.DisplayName, Set.P3.DisplayName, Set.P4.DisplayName)
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
	p1, p2 := c.State.Set.P1, c.State.Set.P2
	if !isTeamA {
		p1, p2 = c.State.Set.P3, c.State.Set.P4
	}
	log.Printf("%s and %s won!\n", p1.DisplayName, p2.DisplayName)
	g := &c.State.Games[len(c.State.Games)-1]
	g.EndTime = time.Now()
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
	teamA := []*goskill.Skill{&c.State.Set.P1.Skill, &c.State.Set.P2.Skill}
	teamB := []*goskill.Skill{&c.State.Set.P3.Skill, &c.State.Set.P4.Skill}
	if isTeamA {
		c.Elo.Rank([][]*goskill.Skill{teamA, teamB})
	} else {
		c.Elo.Rank([][]*goskill.Skill{teamB, teamA})
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
}
