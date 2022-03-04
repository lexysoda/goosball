package api

import (
	"github.com/lexysoda/goosball/model"
	"github.com/slack-go/slack"
)

type Slack struct {
	s *slack.Client
}

func New(token string) *Slack {
	return &Slack{slack.New(token)}
}

func (s *Slack) Send(channel, message string) error {
	text := slack.MsgOptionText(message, false)
	_, _, err := s.s.PostMessage(channel, text)
	return err
}

func (s *Slack) GetUser(id string) (*model.User, error) {
	userProfile, err := s.s.GetUserProfile(&slack.GetUserProfileParameters{
		UserID:        id,
		IncludeLabels: false,
	})
	if err != nil {
		return nil, err
	}
	return &model.User{
		ID:          id,
		RealName:    userProfile.RealName,
		DisplayName: userProfile.DisplayName,
		Avatar:      userProfile.Image512,
	}, nil
}
