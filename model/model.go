package model

import (
	"time"

	"github.com/lexysoda/goskill"
)

type User struct {
	ID          string
	DisplayName string
	RealName    string
	Avatar      string
	Goskill     goskill.Skill
}

func (u User) Skill() *goskill.Skill {
	return &u.Goskill
}

type Set struct {
	ID int64
	P1 User
	P2 User
	P3 User
	P4 User
}

type Game struct {
	ID        int64
	SetID     int64
	GoalsA    int
	GoalsB    int
	StartTime time.Time
	EndTime   time.Time
}
