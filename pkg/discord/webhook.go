package discord

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/wulfixyt/anubis-monitors/pkg/config"
)

type WebhookStruct struct {
	Site    string `json:"website"`
	Product struct {
		Name     string `json:"name"`
		EventId  string `json:"eventId"`
		Venue    string `json:"venue"`
		Date     string `json:"date"`
		IsResale string `json:"isResale"`
		Row      string `json:"row"`
		Section  string `json:"section"`
		Seat     string `json:"seat"`
		Price    string `json:"price"`
		Url      string `json:"url"`
		Image    string `json:"image"`
		Expiry   string `json:"expiry"`
	} `json:"-"`
	Proxy struct {
		Host string
	} `json:"-"`
	IconUrl string
}

type webhook struct {
	Embeds []webhookEmbed `json:"embeds"`
}

type webhookEmbed struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Color     int    `json:"color"`
	Timestamp string `json:"timestamp"`
	Footer    struct {
		IconURL string `json:"icon_url"`
		Text    string `json:"text"`
	} `json:"footer"`
	Thumbnail struct {
		URL string `json:"url"`
	} `json:"thumbnail"`
	Fields []webhookField `json:"fields"`
	Author struct {
		IconURL string `json:"icon_url"`
		Name    string `json:"name"`
	} `json:"author"`
}

type webhookField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// Generating discord timestamp - Fix timestamp
func discordTimestamp() string {
	ts := strings.Replace(time.Now().UTC().Format(time.RFC3339), "Z", ".000Z", -1)
	split := strings.Split(ts, "+")
	if len(split) > 1 {
		return split[0] + ".000Z"
	} else {
		return ts
	}
}

// Webhook for public success logger
func Webhook(task WebhookStruct) {
	wh := webhook{}

	if !strings.Contains(task.Product.Url, "http") {
		task.Product.Url = "https://anubisio.com/"
	}

	embed := webhookEmbed{
		Title:     "Detected Restock",
		Color:     13390566,
		Timestamp: discordTimestamp(),
	}

	if task.Product.Url != "" {
		embed.URL = task.Product.Url
	}

	embed.Footer.IconURL = "https://cdn.discordapp.com/attachments/1159992148419166251/1160242847094677545/center.png?ex=6533f35b&is=65217e5b&hm=661341dcc5ce128e78795f29cf3002bfbb18312a1aad4f3e679ddacbd73839bd&"
	embed.Footer.Text = "AnubisIO"
	embed.Thumbnail.URL = task.Product.Image

	if task.IconUrl != "" {
		embed.Author.Name = task.Site
		embed.Author.IconURL = task.IconUrl
	}

	wh.Embeds = append(wh.Embeds,
		embed,
	)
	wh.Embeds[0].Fields = append(wh.Embeds[0].Fields,
		webhookField{
			Name:  "Product",
			Value: task.Product.Name,
		},
	)

	payload, _ := json.Marshal(wh)

	if config.Config.Config.PublicWebhook != "" {
		res, err := http.Post(config.Config.Config.PublicWebhook, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			time.Sleep(time.Duration(10) * time.Millisecond)
		}

		res.Body.Close()
	}

	res, err := http.Post(config.Config.Config.Webhook, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	defer res.Body.Close()
}
