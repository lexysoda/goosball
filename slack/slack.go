package slack

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Slack struct {
	*slack.Client
	messages chan SlackMessage
}

type SlackMessage struct {
	Sender string
	Text   string
}

func Init() *Slack {
	s := &Slack{Client: slack.New(
		os.Getenv("SLACK_TOKEN"),
		slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
	)}
	go s.Start()
	return s
}

func (s *Slack) Start() {
	client := socketmode.New(s.Client)
	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				log.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				log.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				log.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Ignored %+v\n", evt)

					continue
				}
				client.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.AppMentionEvent:
						m := SlackMessage{ev.User, ev.Text}
						select {
						case s.messages <- m:
						default:
							log.Println("Failed to process message: %+v\n", m)
						}
					default:
						log.Printf("Unexpected Event: %s", innerEvent.Type)
					}
				default:
					log.Println("unsupported Events API event received")
				}
			default:
				log.Printf("Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()
	client.Run()
}

func (s *Slack) RegisterMessageReceiver(c chan SlackMessage) {
	s.messages = c
}
