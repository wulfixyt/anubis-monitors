package eventim

import (
	"client/cookiejar"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	http "github.com/bogdanfinn/fhttp"
	"github.com/wulfixyt/anubis-monitors/pkg/sites/utils"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
)

func session(task *structs.Task) bool {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s/", task.EventimVariables.Authority), nil)

	req.Header = http.Header{
		"sec-ch-ua":                 {utils.GetSecChUa(task.EventimVariables.UserAgent.Web)},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"Windows"`},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {task.EventimVariables.UserAgent.Web},
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-user":            {"?1"},
		"sec-fetch-dest":            {"document"},
		"accept-encoding":           {"gzip, deflate, br"},
		"accept-language":           {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go utils.TaskLog(task, "Session Error - Proxy Error", "red")

		time.Sleep(time.Duration(task.Delay) * time.Millisecond)

		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		task.EventimVariables.Referer = res.Request.URL.String()

		body, _, _ := utils.ParseResponse(res)

		doc, _ := htmlquery.Parse(strings.NewReader(body))
		scripts, err := htmlquery.QueryAll(doc, "//script[@type='text/javascript']")
		if err != nil {
			go utils.TaskLog(task, "Session Error - Failed to parse Response", "red")

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		if len(scripts) == 0 {
			go utils.TaskLog(task, "Session Error - Response Error", "red")

			utils.ChangeRoundtripper(task, task.Client)

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		task.EventimVariables.AkamaiUrl = "https://" + task.EventimVariables.Authority + htmlquery.SelectAttr(scripts[len(scripts)-1], "src")

		// Retrieving Akamai script
		config := getAkamai(task, task.EventimVariables.Referer)

		// Submitting SensorData
		solveAkamai(task, 2, config, 300, 1300, task.EventimVariables.Referer, false, false)
		time.Sleep(2 * time.Second)

		solveAkamai(task, rand.Intn(2)+2, config, 300, 1300, task.EventimVariables.Referer, true, false)
		time.Sleep(4 * time.Second)
		return true
	} else if res.StatusCode == 403 {
		go utils.TaskLog(task, "Session Error - (403)", "red")

		time.Sleep(5 * time.Second)

		task.Jar, _ = cookiejar.New(nil)
		task.Client.Jar = task.Jar

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 429 {
		go utils.TaskLog(task, "Session Error - Rate Limit", "red")

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 503 {
		go utils.TaskLog(task, "Detected Queue", "magenta")

		time.Sleep(30 * time.Second)
		return false
	} else {
		body, _, _ := utils.ParseResponse(res)
		if strings.Contains(body, "Waiting Room page") {
			go utils.TaskLog(task, "Detected Queue", "magenta")

			time.Sleep(30 * time.Second)
			return false
		}

		go utils.TaskLog(task, fmt.Sprintf("Session Error - (%d)", res.StatusCode), "red")
	}

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)
	return false
}

func getAkamai(task *structs.Task, referer string) string {
	req, _ := http.NewRequest("GET", task.EventimVariables.AkamaiUrl, nil)

	req.Header = http.Header{
		"sec-ch-ua":          {utils.GetSecChUa(task.EventimVariables.UserAgent.Web)},
		"sec-ch-ua-mobile":   {"?0"},
		"user-agent":         {task.EventimVariables.UserAgent.Web},
		"sec-ch-ua-platform": {`"Windows"`},
		"accept":             {"*/*"},
		"sec-fetch-site":     {"same-origin"},
		"sec-fetch-mode":     {"no-cors"},
		"sec-fetch-dest":     {"script"},
		"referer":            {referer},
		"accept-encoding":    {"gzip, deflate, br"},
		"accept-language":    {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"sec-ch-ua", "sec-ch-ua-mobile", "user-agent", "sec-ch-ua-platform", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "referer", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	resp, err := task.Client.Do(req)
	if err != nil {
		go utils.TaskLog(task, "Akamai Error - Proxy Error", "red")

		utils.ChangeRoundtripper(task, task.Client)
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)
		return ""
	}

	defer resp.Body.Close()

	_, body, _ := utils.ParseResponse(resp)

	return utils.GetConfig(task, body)
}

func solveAkamai(task *structs.Task, amount int, config string, minSleep int, maxSleep int, referer string, mouseEvents bool, keyboardEvents bool) {
	for i := 0; i < amount; i++ {
		time.Sleep(time.Duration(rand.Intn(maxSleep)+minSleep) * time.Millisecond)

		sensorData := utils.GetSensor(task, referer, utils.GetValue(task, "_abck", task.EventimVariables.Authority), utils.GetValue(task, "bm_sz", task.EventimVariables.Authority), task.EventimVariables.UserAgent.Web, config, mouseEvents, keyboardEvents)
		if len(sensorData) < 200 {
			sensorData = utils.GetSensor(task, referer, utils.GetValue(task, "_abck", task.EventimVariables.Authority), utils.GetValue(task, "bm_sz", task.EventimVariables.Authority), task.EventimVariables.UserAgent.Web, config, mouseEvents, keyboardEvents)
		}

		req, _ := http.NewRequest("POST", task.EventimVariables.AkamaiUrl, strings.NewReader(fmt.Sprintf(`{"sensor_data":"%s"}`, sensorData)))

		req.Header = http.Header{
			"sec-ch-ua":          {utils.GetSecChUa(task.EventimVariables.UserAgent.Web)},
			"sec-ch-ua-platform": {`"Windows"`},
			"sec-ch-ua-mobile":   {"?0"},
			"user-agent":         {task.EventimVariables.UserAgent.Web},
			"content-type":       {"text/plain"},
			"accept":             {"*/*"},
			"origin":             {"https://" + task.EventimVariables.Authority},
			"sec-fetch-site":     {"same-origin"},
			"sec-fetch-mode":     {"cors"},
			"sec-fetch-dest":     {"empty"},
			"referer":            {referer},
			"accept-encoding":    {"gzip, deflate, br"},
			"accept-language":    {"en-US,en;q=0.9"},
			http.HeaderOrderKey: {
				"content-length", "sec-ch-ua", "sec-ch-ua-platform", "sec-ch-ua-mobile", "user-agent", "content-type", "accept", "origin", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "referer", "accept-encoding", "accept-language",
			},
			http.PHeaderOrderKey: {
				":method", ":authority", ":scheme", ":path",
			},
		}

		resp, err := task.Client.Do(req)
		if err != nil {
			utils.ChangeRoundtripper(task, task.Client)
			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return
		}

		resp.Body.Close()

		for _, value := range resp.Cookies() {
			if value.Name == "_abck" {
				if strings.Contains(value.Value, "~0~") {
					return
				}
			}
		}
	}
}
