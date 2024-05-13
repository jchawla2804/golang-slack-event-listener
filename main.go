package main

import (
	"context"
	"log"
	"os"

	"github.com/jchawla2804/golang-slack-event-listener/events"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error Loading Property file")
	}

	slackClient := slack.New(
		os.Getenv("SLACK_BOT_TOKEN"),
		//slack.OptionDebug(true),
		slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
	)

	socketClient := socketmode.New(
		slackClient,
		//socketmode.OptionDebug(true),
		//socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	log.Println("Connectivity successful")

	Context, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting Down listener")
				return

			case event := <-socketClient.Events:
				switch event.Type {
				case socketmode.EventTypeEventsAPI:
					log.Println("Slack api event")
					eventApiEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could Not typecast the event : %v\n", event)
						continue
					}

					socketClient.Ack(*event.Request)
					err = events.HandleSlackEventMessage(eventApiEvent, client)
					if err != nil {
						log.Fatal(err)
					}

				case socketmode.EventTypeSlashCommand:
					slackCommandEvent, ok := event.Data.(slack.SlashCommand)
					if !ok {
						log.Printf("Could Not typecast the event %v\n", slackCommandEvent)
						continue
					}

					socketClient.Ack(*event.Request)
					err := events.HandleSlackCommands(slackClient, slackCommandEvent)
					if err != nil {
						log.Fatal(err)
					}
				case socketmode.EventTypeInteractive:
					callbackEvent, ok := event.Data.(slack.InteractionCallback)
					log.Println("Message Recieved")
					if !ok {
						log.Fatal("Cound not typecast event")
					}

					switch callbackEvent.Type {

					// case for block actions
					case slack.InteractionTypeBlockActions:
						socketClient.Ack(*event.Request)

						actiontype := callbackEvent.ActionCallback.BlockActions[0].Type

						switch actiontype {
						case slack.ActionType(slack.OptTypeStatic):
							err := events.HandlePlatformInformation(slackClient, callbackEvent.ActionCallback.BlockActions[0].SelectedOption.Text.Text, callbackEvent.ActionCallback.BlockActions[0].SelectedOption.Value)
							if err != nil {
								log.Fatal(err.Error())
							}

						default:
							err = events.HandleInteractiveDialogBoxEvent(slackClient, callbackEvent.ActionCallback.BlockActions[0].Value, callbackEvent.TriggerID)
							if err != nil {
								log.Fatal(err.Error())
							}

						}

					// case Submission events
					case slack.InteractionTypeViewSubmission:
						socketClient.Ack(*event.Request)
						var username, password, typeOfAuth string

						if callbackEvent.View.State.Values["username"]["user"].Value == "" {
							username = callbackEvent.View.State.Values["clientId"]["clientId"].Value
							password = callbackEvent.View.State.Values["clientSecret"]["clientSecret"].Value
							typeOfAuth = "oauth"
						} else {
							username = callbackEvent.View.State.Values["username"]["user"].Value
							password = callbackEvent.View.State.Values["password"]["pass"].Value
							typeOfAuth = "basic-auth"
						}

						err = events.HandleLogin(slackClient, username, password, typeOfAuth)
						if err != nil {
							log.Println(err.Error())
						}
					}

				}

			}
		}
	}(Context, slackClient, socketClient)

	socketClient.Run()

}
