package bot

import (
	"log"
	"os"
	"regexp"

	"github.com/lexysoda/goosball/controller"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Slack struct {
	client     *socketmode.Client
	controller *controller.Controller
}

var commandRegex = regexp.MustCompile(`^<@(\w+)>\s*(\w+)\s*(.*)$`)
var usersRegex = regexp.MustCompile(`<@(\w+)>`)

func New(c *controller.Controller) *Slack {
	s := &Slack{
		client: socketmode.New(slack.New(
			os.Getenv("SLACK_TOKEN"),
			slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
		)),
		controller: c,
	}
	s.start()
	return s
}

func (s *Slack) start() {
	go func() {
		for evt := range s.client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				log.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				log.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				log.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeHello:
				log.Println("Got hello event.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Ignored %+v\n", evt)
					continue
				}
				s.client.Ack(*evt.Request)
				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.AppMentionEvent:
						s.handleMessage(ev.User, ev.Text)
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
	go s.client.Run()
}

func (s *Slack) handleMessage(user, text string) {
	cmd := commandRegex.FindStringSubmatch(text)
	if cmd == nil {
		return
	}
	switch cmd[2] {
	case "play":
		s.controller.AddToQueue(user)
	case "queue":
		s.controller.SendQueueSlack()
	case "cancel":
		s.controller.CancelSet()
	case "add":
		matches := usersRegex.FindAllStringSubmatch(cmd[3], -1)
		if matches == nil {
			return
		}
		users := []string{}
		for _, m := range matches {
			users = append(users, m[1])
		}
		s.controller.AddToQueue(users...)
	case "remove":
		s.controller.RemoveFromQueue(user)
	default:
		log.Println("Unknown command")
	}
}
