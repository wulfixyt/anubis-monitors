package fansale

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/wulfixyt/anubis-monitors/pkg/akamai"
	"github.com/wulfixyt/anubis-monitors/pkg/log"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
	"github.com/wulfixyt/anubis-monitors/pkg/utils"
)

func checkExpiry(id int) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := webhookHandler[id]; !ok {
		webhookHandler[id] = true
		return true
	}

	return false
}

func getAkamai(task *structs.Task, referer string) string {
	req, _ := http.NewRequest("GET", task.FansaleVariables.AkamaiUrl, nil)

	req.Header = http.Header{
		"Connection":         {"keep-alive"},
		"sec-ch-ua":          {utils.GetSecChUa(task.FansaleVariables.UserAgent)},
		"sec-ch-ua-mobile":   {"?0"},
		"User-Agent":         {task.FansaleVariables.UserAgent},
		"sec-ch-ua-platform": {`"Windows"`},
		"Accept":             {"*/*"},
		"Sec-Fetch-Site":     {"same-origin"},
		"Sec-Fetch-Mode":     {"no-cors"},
		"Sec-Fetch-Dest":     {"script"},
		"Referer":            {referer},
		"Accept-Encoding":    {"gzip, deflate, br"},
		"Accept-Language":    {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"host", "connection", "sec-ch-ua", "sec-ch-ua-mobile", "user-agent", "sec-ch-ua-platform", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "referer", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	resp, err := task.Client.Do(req)
	if err != nil {
		go log.ErrorLogger.Println(log.Format(task, "Akamai Error - Proxy Error", "red"))

		utils.ChangeRoundtripper(task, task.Client)
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)
		return ""
	}

	defer resp.Body.Close()

	_, body, _ := utils.ParseResponse(resp)

	return akamai.GetConfig(task, body)
}

func solveAkamai(task *structs.Task, amount int, config string, minSleep int, maxSleep int, referer string, mouseEvents bool, keyboardEvents bool) {
	for i := 0; i < amount; i++ {
		time.Sleep(time.Duration(rand.Intn(maxSleep)+minSleep) * time.Millisecond)

		sensorData := akamai.GetSensor(task, referer, utils.GetValue(task, "_abck", task.FansaleVariables.Authority), utils.GetValue(task, "bm_sz", task.FansaleVariables.Authority), task.FansaleVariables.UserAgent, config, mouseEvents, keyboardEvents)
		if len(sensorData) < 200 {
			sensorData = akamai.GetSensor(task, referer, utils.GetValue(task, "_abck", task.FansaleVariables.Authority), utils.GetValue(task, "bm_sz", task.FansaleVariables.Authority), task.FansaleVariables.UserAgent, config, mouseEvents, keyboardEvents)
		}

		req, _ := http.NewRequest("POST", task.FansaleVariables.AkamaiUrl, strings.NewReader(fmt.Sprintf(`{"sensor_data":"%s"}`, sensorData)))

		req.Header = http.Header{
			"Connection":         {"keep-alive"},
			"sec-ch-ua":          {utils.GetSecChUa(task.FansaleVariables.UserAgent)},
			"sec-ch-ua-platform": {`"Windows"`},
			"sec-ch-ua-mobile":   {"?0"},
			"User-Agent":         {task.FansaleVariables.UserAgent},
			"Content-Type":       {"text/plain"},
			"Accept":             {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
			"Origin":             {"https://" + task.FansaleVariables.Authority},
			"Sec-Fetch-Site":     {"same-origin"},
			"Sec-Fetch-Mode":     {"cors"},
			"Sec-Fetch-Dest":     {"empty"},
			"Referer":            {referer},
			"Accept-Encoding":    {"gzip, deflate, br"},
			"Accept-Language":    {"en-US,en;q=0.9"},
			http.HeaderOrderKey: {
				"host", "connection", "content-length", "sec-ch-ua", "sec-ch-ua-platform", "sec-ch-ua-mobile", "user-agent", "content-type", "accept", "origin", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "referer", "accept-encoding", "accept-language",
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

		fmt.Println(resp.Cookies())

		for _, value := range resp.Cookies() {
			if value.Name == "_abck" {
				if strings.Contains(value.Value, "~0~") {
					return
				}
			}
		}
	}
}
