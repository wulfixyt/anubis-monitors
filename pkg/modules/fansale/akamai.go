package fansale

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	http "github.com/bogdanfinn/fhttp"
	"github.com/wulfixyt/anubis-monitors/pkg/log"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
	"github.com/wulfixyt/anubis-monitors/pkg/utils"
)

func session(task *structs.Task) bool {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s/fansale/", task.FansaleVariables.Authority), nil)

	req.Header = http.Header{
		"Connection":                {"keep-alive"},
		"sec-ch-ua":                 {utils.GetSecChUa(task.FansaleVariables.UserAgent)},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"Windows"`},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {task.FansaleVariables.UserAgent},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Sec-Fetch-Site":            {"none"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-User":            {"?1"},
		"Sec-Fetch-Dest":            {"document"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"host", "connection", "sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go log.ErrorLogger.Println(log.Format(task, "Session Error - Proxy Error", "red"))

		time.Sleep(time.Duration(task.Delay) * time.Millisecond)

		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		task.FansaleVariables.Referer = res.Request.URL.String()

		body, _, _ := utils.ParseResponse(res)

		doc, _ := htmlquery.Parse(strings.NewReader(body))
		scripts, err := htmlquery.QueryAll(doc, "//script[@type='text/javascript']")
		if err != nil {
			go log.ErrorLogger.Println(log.Format(task, "Session Error - Failed to parse Response", "red"))

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		if len(scripts) == 0 {
			go log.ErrorLogger.Println(log.Format(task, "Session Error - Response Error", "red"))

			utils.ChangeRoundtripper(task, task.Client)

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		task.FansaleVariables.AkamaiUrl = "https://" + task.FansaleVariables.Authority + htmlquery.SelectAttr(scripts[len(scripts)-1], "src")

		// Retrieving Akamai script
		task.FansaleVariables.AkamaiConfig = getAkamai(task, task.FansaleVariables.Referer)

		// Submitting SensorData
		solveAkamai(task, 2, task.FansaleVariables.AkamaiConfig, 300, 1300, task.FansaleVariables.Referer, false, false)
		time.Sleep(2 * time.Second)

		solveAkamai(task, rand.Intn(2)+2, task.FansaleVariables.AkamaiConfig, 300, 1300, task.FansaleVariables.Referer, true, false)
		time.Sleep(4 * time.Second)
		return true
	} else if res.StatusCode == 403 {
		go log.ErrorLogger.Println(log.Format(task, "Session Error - (403)", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 429 {
		go log.ErrorLogger.Println(log.Format(task, "Session Error - Rate Limit", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 503 {
		go log.ErrorLogger.Println(log.Format(task, "Detected Queue", "magenta"))

		time.Sleep(30 * time.Second)
		return false
	} else {
		body, _, _ := utils.ParseResponse(res)
		if strings.Contains(body, "Waiting Room page") {
			go log.ErrorLogger.Println(log.Format(task, "Detected Queue", "magenta"))

			time.Sleep(30 * time.Second)
			return false
		}

		go log.ErrorLogger.Println(log.Format(task, fmt.Sprintf("Session Error - (%d)", res.StatusCode), "red"))
	}

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)
	return false
}