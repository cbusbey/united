package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/cbusbey/united"
)

type shipmentConfig struct {
	Name        string
	OrderNumber string
	NickName    string
}

type attachment struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

type payload struct {
	Text        string       `json:"text"`
	Username    string       `json:"username"`
	IconEmoji   string       `json:"icon_emoji"`
	Channel     string       `json:"channel,omitempty"`
	Attachments []attachment `json:"attachments"`
}

func toPayload(config shipmentConfig, tracking united.ShipmentTrackingResponse) (payload, error) {
	van := tracking.Shipment.Van

	gMapsURL := fmt.Sprintf("https://www.google.com/maps/place/%v+%v",
		van.LastReportedCity, van.LastReportedState)

	milesAway, err := strconv.Atoi(van.TrackingMiles)
	if err != nil {
		return payload{}, err
	}

	var text string

	if milesAway > 0 {
		text = fmt.Sprintf("%v is %v miles %v of <%v | %v, %v> %v",
			config.NickName,
			milesAway,
			van.TrackingDirection,
			gMapsURL,
			van.LastReportedCity,
			van.LastReportedState,
			van.LastReportedDateTime)
	} else {
		text = fmt.Sprintf("%v at <%v | %v, %v> %v",
			config.NickName,
			gMapsURL,
			van.LastReportedCity,
			van.LastReportedState,
			van.LastReportedDateTime)
	}

	return payload{
		Channel:   *channel,
		Username:  "United Tracker",
		IconEmoji: ":round_pushpin:",
		Attachments: []attachment{
			{
				Color: "good",
				Text:  text,
			},
		},
	}, nil
}

var webhookURL = flag.String("webhookurl", "", "Slack WebHook URL")
var duration = flag.Duration("duration", 15*time.Minute, "Time between lookups")
var channel = flag.String("channel", "", "Channel to send to, defaults to that configured for the webhook url")
var nickname = flag.String("nickname", "", "Nickname for the shipment to track")
var name = flag.String("name", "", "Last name on the United order")
var order = flag.String("order", "", "United order number")
var test = flag.Bool("test", false, "Set to true to just generate a message but do not send")

func flagMustNotBeEmpty(flagValue *string) {
	if flagValue == nil || *flagValue == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	flagMustNotBeEmpty(webhookURL)
	flagMustNotBeEmpty(name)
	flagMustNotBeEmpty(nickname)
	flagMustNotBeEmpty(order)

	shipConfig := shipmentConfig{
		Name:        *name,
		OrderNumber: *order,
		NickName:    *nickname,
	}

	var lastUpdate united.ShipmentTrackingResponse
	ticker := time.Tick(*duration)

	for {
		tracking, err := united.GetTrackingUpdate(shipConfig.Name, shipConfig.OrderNumber)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(tracking, lastUpdate) {
			p, err := toPayload(shipConfig, tracking)
			if err != nil {
				panic(err)
			}

			b, err := json.Marshal(p)
			if err != nil {
				panic(err)
			}

			if *test {
				fmt.Printf("%s\n", b)
				os.Exit(0)
			}

			if _, err := http.Post(*webhookURL, "text/json", bytes.NewReader(b)); err != nil {
				panic(err)
			}

			lastUpdate = tracking
		}

		<-ticker
	}
}
