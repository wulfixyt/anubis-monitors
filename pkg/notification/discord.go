package notification

import (
	"fmt"
	"net/http"
	"strings"
)

func SendDiscord(website string, event string, url string, content string) {
	webhook := "https://discord.com/api/webhooks/1166460594942136386/aUm0poROZAYAcMQI4aYWCu5DG-egkgZfWOVfEfsZZVzJWBMeYV9nowlCWI4CWd-9YUJd"
	payload := fmt.Sprintf(`{"embeds":[{"title":"Restock detected","description":"%s","color":13390566,"fields":[{"name":"Website","value":"%s"},{"name":"Event","value":"[%s](%s)"}],"footer":{"text":"AnubisIO","icon_url":"https://cdn.discordapp.com/attachments/1159992148419166251/1160242847094677545/center.png?ex=6533f35b&is=65217e5b&hm=661341dcc5ce128e78795f29cf3002bfbb18312a1aad4f3e679ddacbd73839bd&"}}]}`, strings.ReplaceAll(content, `"`, `\"`), website, event, url)

	go func() {
		res, err := http.Post(webhook, "application/json", strings.NewReader(payload))
		if err != nil {
			return
		}

		defer res.Body.Close()
	}()
}
