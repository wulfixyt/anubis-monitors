package socket

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/wulfixyt/anubis-monitors/pkg/config"
)

type monitorResponse struct {
	Success bool `json:"success"`
	Message []struct {
		Website  string   `json:"website"`
		Products []string `json:"products"`
	} `json:"message"`
}

func loadSettings() error {
	file, err := os.Open("./settings.json")
	if err != nil {
		return err
	}

	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, &config.Config)

	return err
}

func formatWebsite(website string) string {
	switch {
	case strings.Contains(strings.ToUpper(website), "EVENTIM"):
		return "EVENTIM"
	case strings.Contains(strings.ToUpper(website), "FANSALE"):
		return "FANSALE"
	case strings.Contains(strings.ToUpper(website), "TICKETMASTER"):
		return "TICKETMASTER"
	default:
		return ""
	}
}
