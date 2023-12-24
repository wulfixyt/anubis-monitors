package eventim

import (
	"encoding/json"
	"fmt"
	"github.com/wulfixyt/anubis-monitors/pkg/notification"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"client"
	"client/cookiejar"

	"github.com/antchfx/htmlquery"
	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/wulfixyt/anubis-monitors/pkg/discord"
	"github.com/wulfixyt/anubis-monitors/pkg/log"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
	"github.com/wulfixyt/anubis-monitors/pkg/utils"
)

func Run(task *structs.Task) {
	defer func() {
		if r := recover(); r != nil {
			log.ErrorLogger.Println(log.Format(task, fmt.Sprintf("Recovered from panic: %v", r), "red"))
			time.Sleep(5 * time.Second)

			Run(task)
			return
		}
	}()

	if strings.Contains(task.Input, "|") {
		task.Code = strings.Split(task.Input, "|")[1]
		task.Input = strings.Split(task.Input, "|")[0]
	}

	if strings.Contains(task.Input, "http") {
		if !strings.Contains(task.Input, "https") {
			task.Input = strings.ReplaceAll(task.Input, "http", "https")
		}

		// Detects all numbers from the url
		re := regexp.MustCompile(`[^\d]`)
		task.Event.EventId = re.ReplaceAllString(strings.Split(strings.ReplaceAll(strings.ReplaceAll(task.Input, "2024-", ""), "2025-", ""), "?")[0], "")

		u, _ := url.Parse(task.Input)
		task.EventimVariables.AffiliateId = u.Query().Get("affiliate")
	} else if strings.Contains(task.Input, "www.") {
		task.Input = "https://" + task.Input
		// Detects all numbers from the url
		re := regexp.MustCompile(`[^\d]`)
		task.Event.EventId = re.ReplaceAllString(strings.Split(task.Input, "?")[0], "")

		u, _ := url.Parse(task.Input)
		task.EventimVariables.AffiliateId = u.Query().Get("affiliate")
	} else {
		task.Event.EventId = task.Input
		task.EventimVariables.AffiliateId = ""
	}

	task.EventimVariables.Authority = "www.eventim.de"
	if strings.Contains(strings.ToLower(task.Site), "-ch") || strings.Contains(task.Input, "ticketcorner.ch") {
		task.EventimVariables.Authority = "www.ticketcorner.ch"
	} else if strings.Contains(strings.ToLower(task.Site), "-at") || strings.Contains(task.Input, "oeticket.com") {
		task.EventimVariables.Authority = "www.oeticket.com"
	} else if strings.Contains(strings.ToLower(task.Site), "-it") || strings.Contains(task.Input, "ticketone.it") {
		task.EventimVariables.Authority = "www.ticketone.it"
	}

	rand.Seed(time.Now().UnixNano())
	utils.RotateProxy(task)

	rt, err := client.NewRoundtripper(profiles.Chrome_110, client.Settings{Proxy: task.Proxy})
	if err != nil {
		return
	}

	task.Jar, _ = cookiejar.New(nil)

	task.Client = &http.Client{
		Transport: rt,
		Jar:       task.Jar,
		Timeout:   20 * time.Second,
	}

	go log.InfoLogger.Println(log.Format(task, "Generating Session", "white"))

	task.EventimVariables.UserAgent.Web = utils.GetUserAgent(task)

	for !session(task) {
		select {
		case <-task.Ctx.Done():
			return
		default:
			continue
		}
	}

	go log.InfoLogger.Println(log.Format(task, "Getting Product Page", "white"))

	task.Event.Url = task.Input
	task.EventimVariables.Referer = task.Event.Url

	for !monitorReload(task, "session") {
		select {
		case <-task.Ctx.Done():
			return
		default:
			continue
		}
	}

	if task.EventimVariables.RequiresPromo && task.Code != "" {
		go log.InfoLogger.Println(log.Format(task, "Adding Access Code", "white"))

		for !enterCode(task) {
			select {
			case <-task.Ctx.Done():
				return
			default:
				continue
			}
		}

		task.EventimVariables.SessionStart = time.Now().Unix()

		for !monitor(task) {
			select {
			case <-task.Ctx.Done():
				return
			default:
				if time.Now().Unix()-task.EventimVariables.SessionStart > 1800 {
					go log.InfoLogger.Println(log.Format(task, "Adding Access Code", "white"))

					for !enterCode(task) {
						select {
						case <-task.Ctx.Done():
							return
						default:
							continue
						}
					}

					task.EventimVariables.SessionStart = time.Now().Unix()
				}

				continue
			}
		}

		for !monitorReload(task, "code") {
			select {
			case <-task.Ctx.Done():
				return
			default:
				if time.Now().Unix()-task.EventimVariables.SessionStart > 1800 {
					go log.InfoLogger.Println(log.Format(task, "Adding Access Code", "white"))

					for !enterCode(task) {
						select {
						case <-task.Ctx.Done():
							return
						default:
							continue
						}
					}

					task.EventimVariables.SessionStart = time.Now().Unix()
				}

				continue
			}
		}
	}

}

func enterCode(task *structs.Task) bool {
	//utils.DelCookie(task.Client, task.EventimVariables.Authority, "_abck")
	//utils.DelCookie(task.Client, task.EventimVariables.Authority, "ak_bmsc")
	//utils.DelCookie(task.Client, task.EventimVariables.Authority, "bm_sv")
	//utils.DelCookie(task.Client, task.EventimVariables.Authority, "bm_sz")
	//utils.DelCookie(task.Client, task.EventimVariables.Authority, "akawr")

	data := fmt.Sprintf(`{"evId":%s,"promoId":%s,"promotionCodes":[{"promotionCode":"%s","id":0}]}`, task.Event.EventId, task.EventimVariables.PromoCode, task.Code)

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://%s/api/promocode/;%5c.js?affiliate=%s&force_session=true", task.EventimVariables.Authority, task.EventimVariables.AffiliateId), strings.NewReader(data))

	req.Header = http.Header{
		"sec-ch-ua":          {utils.GetSecChUa(task.EventimVariables.UserAgent.Web)},
		"accept":             {"application/json, text/javascript, */*; q=0.01"},
		"content-type":       {"application/json"},
		"x-requested-with":   {"XMLHttpRequest"},
		"sec-ch-ua-mobile":   {"?0"},
		"user-agent":         {task.EventimVariables.UserAgent.Web},
		"sec-ch-ua-platform": {`"Windows"`},
		"origin":             {"https://" + task.EventimVariables.Authority},
		"sec-fetch-site":     {"same-origin"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-dest":     {"empty"},
		"referer":            {task.EventimVariables.Referer},
		"accept-encoding":    {"gzip, deflate, br"},
		"accept-language":    {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"content-length", "sec-ch-ua", "accept", "content-type", "x-requested-with", "sec-ch-ua-mobile", "user-agent", "sec-ch-ua-platform", "origin", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "referer", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go log.InfoLogger.Println(log.Format(task, "Code Error - Proxy Error", "red"))

		time.Sleep(time.Duration(task.Delay) * time.Millisecond)

		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "_abck")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "ak_bmsc")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "bm_sv")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "bm_sz")

	if res.StatusCode == 200 {
		_, body, _ := utils.ParseResponse(res)

		type responseStruct struct {
			Token string `json:"token"`
		}

		var response responseStruct

		json.Unmarshal(body, &response)

		task.EventimVariables.Token = response.Token
		if len(task.EventimVariables.Token) > 0 {
			return true
		}
	} else if res.StatusCode == 403 {
		go log.InfoLogger.Println(log.Format(task, fmt.Sprintf("Code Error - %d", res.StatusCode), "red"))

		time.Sleep(5 * time.Second)

		task.Jar, _ = cookiejar.New(nil)
		task.Client.Jar = task.Jar

		utils.ChangeRoundtripper(task, task.Client)

		for !session(task) {
			select {
			case <-task.Ctx.Done():
				return false
			default:
				continue
			}
		}

		for !monitorReload(task, "session") {
			select {
			case <-task.Ctx.Done():
				return false
			default:
				continue
			}
		}
	}

	go log.InfoLogger.Println(log.Format(task, fmt.Sprintf("Code Error - %d", res.StatusCode), "red"))

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)

	utils.ChangeRoundtripper(task, task.Client)
	return false
}

func monitor(task *structs.Task) bool {
	req, _ := http.NewRequest("POST", task.Input, strings.NewReader(`token=`+task.EventimVariables.Token))

	req.Header = http.Header{
		"sec-ch-ua":                 {utils.GetSecChUa(task.EventimVariables.UserAgent.Web)},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"Windows"`},
		"upgrade-insecure-requests": {"1"},
		"origin":                    {"https://" + task.EventimVariables.Authority},
		"content-type":              {"application/x-www-form-urlencoded"},
		"user-agent":                {task.EventimVariables.UserAgent.Web},
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"sec-fetch-site":            {"same-origin"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-user":            {"?1"},
		"sec-fetch-dest":            {"document"},
		"accept-encoding":           {"gzip, deflate, br"},
		"accept-language":           {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"content-length",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"upgrade-insecure-requests",
			"origin",
			"content-type",
			"user-agent",
			"accept",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-user",
			"sec-fetch-dest",
			"accept-encoding",
			"accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go log.InfoLogger.Println(log.Format(task, "Monitor Error - Proxy Error", "red"))

		time.Sleep(time.Duration(task.Delay) * time.Millisecond)

		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "_abck")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "ak_bmsc")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "bm_sv")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "bm_sz")

	if res.StatusCode == 200 {
		body, _, _ := utils.ParseResponse(res)

		doc, err := htmlquery.Parse(strings.NewReader(body))
		if err != nil {
			go log.InfoLogger.Println(log.Format(task, "Parsing Error", "red"))

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		if task.Event.Name == "" {
			nodes := htmlquery.Find(doc, "//h1[@class]")

			task.Event.Name = htmlquery.InnerText(nodes[0])
			task.Event.Url = res.Request.URL.String()
		}

		nodes, err := htmlquery.QueryAll(doc, "//div[@data-max]")
		if err != nil {
			go log.InfoLogger.Println(log.Format(task, "Parsing Error", "red"))

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		for _, node := range nodes {
			maxQty := htmlquery.SelectAttr(node, "data-max")
			if maxQty != "0" {
				go log.InfoLogger.Println(log.Format(task, "Sending Webhook", "green"))

				webhook := discord.WebhookStruct{
					Site:    "Eventim",
					Product: task.Event,
				}

				go discord.Webhook(webhook)

				time.Sleep(time.Duration(task.Delay) * time.Millisecond)
				return true
			}
		}

		go log.InfoLogger.Println(log.Format(task, "Waiting for Tickets", "white"))

		time.Sleep(time.Duration(task.Delay) * time.Millisecond)
		return true
	} else if res.StatusCode == 403 {
		go log.InfoLogger.Println(log.Format(task, "Monitor Error - (403)", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 429 {
		go log.InfoLogger.Println(log.Format(task, "Monitor Error - Rate Limit", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 503 {
		go log.InfoLogger.Println(log.Format(task, "Detected Queue", "magenta"))

		time.Sleep(30 * time.Second)
		return false
	} else {
		go log.InfoLogger.Println(log.Format(task, fmt.Sprintf("Monitor Error - (%d)", res.StatusCode), "red"))
	}

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)
	return false
}

func monitorReload(task *structs.Task, mode string) bool {
	req, _ := http.NewRequest("GET", task.Event.Url, nil)

	req.Header = http.Header{
		"sec-ch-ua":                 {utils.GetSecChUa(task.EventimVariables.UserAgent.Web)},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"Windows"`},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {task.EventimVariables.UserAgent.Web},
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-user":            {"?1"},
		"sec-fetch-dest":            {"document"},
		"referer":                   {task.EventimVariables.Referer},
		"accept-encoding":           {"gzip, deflate, br"},
		"accept-language":           {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"upgrade-insecure-requests",
			"user-agent",
			"accept",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-user",
			"sec-fetch-dest",
			"referer",
			"accept-encoding",
			"accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go log.InfoLogger.Println(log.Format(task, "Monitor Error - Proxy Error", "red"))

		time.Sleep(time.Duration(task.Delay) * time.Millisecond)

		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "_abck")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "ak_bmsc")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "bm_sv")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "bm_sz")
	////utils.DelCookie(task.Client, task.EventimVariables.Authority, "akawr")

	task.Event.Url = res.Request.URL.String()

	if res.StatusCode == 200 {
		body, _, _ := utils.ParseResponse(res)

		doc, err := htmlquery.Parse(strings.NewReader(body))
		if err != nil {
			go log.InfoLogger.Println(log.Format(task, "Parsing Error", "red"))

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		if mode == "session" && strings.Contains(body, `<div class="add-promo-code" style="display: none;">`) {
			task.EventimVariables.PromoCode = "124648"
			if strings.Contains(body, `"activePromoId": `) {
				task.EventimVariables.PromoCode = strings.Split(strings.Split(body, `"activePromoId": `)[1], ",")[0]
			}

			task.EventimVariables.RequiresPromo = true

			scripts, err := htmlquery.QueryAll(doc, "//script[@type='text/javascript']")
			if err != nil || len(scripts) == 0 {
				return true
			}

			task.EventimVariables.Referer = task.Event.Url

			// Retrieving Akamai script
			config := getAkamai(task, task.EventimVariables.Referer)

			// Submitting SensorData
			solveAkamai(task, 2, config, 300, 1300, task.EventimVariables.Referer, false, false)
			time.Sleep(2 * time.Second)

			solveAkamai(task, 2, config, 300, 1300, task.EventimVariables.Referer, true, false)
			time.Sleep(2 * time.Second)

			return true
		}

		if mode != "session" && !strings.Contains(body, `<div class="card-section promotion-card card-grid js-promo-active">`) {
			go log.InfoLogger.Println(log.Format(task, "Monitor Error - Token expired", "red"))

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)

			task.EventimVariables.Referer = task.Event.Url

			// Retrieving Akamai script
			config := getAkamai(task, task.EventimVariables.Referer)

			solveAkamai(task, 2, config, 300, 1300, task.EventimVariables.Referer, false, false)
			time.Sleep(2 * time.Second)

			solveAkamai(task, 1, config, 300, 1300, task.EventimVariables.Referer, true, false)

			for !enterCode(task) {
				select {
				case <-task.Ctx.Done():
					return false
				default:
					continue
				}
			}

			for !monitor(task) {
				select {
				case <-task.Ctx.Done():
					return false
				default:
					continue
				}
			}

			return false
		}

		if task.Event.Name == "" {
			nodes := htmlquery.Find(doc, "//h1[@class]")

			task.Event.Name = htmlquery.InnerText(nodes[0])
			task.Event.Url = res.Request.URL.String()
		}

		nodes, err := htmlquery.QueryAll(doc, "//div[@data-max]")
		if err != nil {
			go log.InfoLogger.Println(log.Format(task, "Parsing Error", "red"))

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		for _, node := range nodes {
			maxQty := htmlquery.SelectAttr(node, "data-max")
			if maxQty != "0" {
				go log.InfoLogger.Println(log.Format(task, "Sending Webhook", "green"))

				notification.SendDiscord("Eventim", task.FansaleVariables.EventId, task.Input, "")

				time.Sleep(time.Duration(task.Delay) * time.Millisecond)
				return false
			}
		}

		go log.InfoLogger.Println(log.Format(task, "Waiting for Tickets", "white"))
	} else if res.StatusCode == 403 {
		go log.InfoLogger.Println(log.Format(task, "Monitor Error - (403)", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 429 {
		go log.InfoLogger.Println(log.Format(task, "Monitor Error - Rate Limit", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 503 {
		go log.InfoLogger.Println(log.Format(task, "Detected Queue", "magenta"))

		time.Sleep(30 * time.Second)
		return false
	} else {
		go log.InfoLogger.Println(log.Format(task, fmt.Sprintf("Monitor Error - (%d)", res.StatusCode), "red"))
	}

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)
	return false
}
