package slack

import (
	"os"

	"github.com/lexysoda/goosball/model"
	"github.com/slack-go/slack"
)

type Slack struct {
	*slack.Client
}

func Init() *Slack {
	return &Slack{slack.New(os.Getenv("SLACK_TOKEN"))}
}

func (s *Slack) GetUser(id string) (*model.User, error) {
	userProfile, err := s.GetUserProfile(&slack.GetUserProfileParameters{
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
