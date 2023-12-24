package config

var Config Settings

type Settings struct {
	Config struct {
		// Private webhook
		Webhook string `json:"webhook"`

		// Public webhook
		PublicWebhook string `json:"publicWebhook"`

		// Delays per website
		Delays map[string]int `json:"delays"`
	} `json:"config"`
}
