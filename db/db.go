package db

import (
	"database/sql"
	"errors"

	"github.com/lexysoda/goosball/model"
	_ "modernc.org/sqlite"
)

var NoRow = errors.New("No result found")

type DB struct {
	*sql.DB
}

func New() (DB, error) {
	db, err := sql.Open("sqlite", "goosball.db")
	if err != nil {
		return DB{}, err
	}
	return DB{db}, nil
}

func (db *DB) AddUser(u *model.User) error {
	_, err := db.Exec(
		"INSERT INTO users (ID, DisplayName, RealName, Avatar, Mu, SigmaSq) VALUES ($1, $2, $3, $4, $5, $6)",
		u.ID,
		u.DisplayName,
		u.RealName,
		u.Avatar,
		u.Skill.Mu,
		u.Skill.SigSq,
	)
	return err
}

func (db *DB) UpdateUser(u *model.User) error {
	_, err := db.Exec(
		"UPDATE users SET DisplayName = $2, RealName = $3, Avatar = $4, Mu = $5, SigmaSq = $6 WHERE ID = $1",
		u.ID,
		u.DisplayName,
		u.RealName,
		u.Avatar,
		u.Skill.Mu,
		u.Skill.SigSq,
	)
	return err
}

func (db *DB) GetUser(id string) (*model.User, error) {
	uGot := &model.User{}
	err := db.QueryRow("SELECT * from users WHERE ID = ?", id).Scan(
		&uGot.ID,
		&uGot.DisplayName,
		&uGot.RealName,
		&uGot.Avatar,
		&uGot.Skill.Mu,
		&uGot.Skill.SigSq,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, NoRow
	}
	return uGot, err
}

func (db *DB) GetUsers() ([]model.User, error) {
	users := []model.User{}
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		u := model.User{}
		if err := rows.Scan(
			&u.ID,
			&u.DisplayName,
			&u.RealName,
			&u.Avatar,
			&u.Skill.Mu,
			&u.Skill.SigSq,
		); err != nil {
			return users, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (db *DB) GetSet(id int64) (*model.Set, error) {
	s := &model.Set{}
	err := db.QueryRow("SELECT * from sets WHERE ID = ?", id).Scan(
		&s.ID,
		&s.P1.ID,
		&s.P2.ID,
		&s.P3.ID,
		&s.P4.ID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, NoRow
	}
	return s, err
}

func (db *DB) NewSet(s *model.Set) (int64, error) {
	res, err := db.Exec("INSERT INTO sets (P1, P2, P3, P4) VALUES ($1, $2, $3, $4)", s.P1.ID, s.P2.ID, s.P3.ID, s.P4.ID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) NewGame(g *model.Game) (int64, error) {
	res, err := db.Exec("INSERT INTO games (SetID, GoalsA, GoalsB, StartTime, EndTime) VALUES ($1, $2, $3, $4, $5)", g.SetID, g.GoalsA, g.GoalsB, g.StartTime, g.EndTime)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
