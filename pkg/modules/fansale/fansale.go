package fansale

import (
	"client"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/wulfixyt/anubis-monitors/pkg/log"
	"github.com/wulfixyt/anubis-monitors/pkg/notification"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
	"github.com/wulfixyt/anubis-monitors/pkg/utils"
)

var (
	webhookHandler = make(map[int]bool)
	webhookArray   = []int{}
	mutex          sync.RWMutex
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

	utils.RotateProxy(task)

	rt, err := client.NewRoundtripper(profiles.Chrome_117, client.Settings{Proxy: task.Proxy})
	if err != nil {
		return
	}

	task.Jar, _ = cookiejar.New(nil)

	task.Client = &http.Client{
		Transport: rt,
		//Jar:       task.Jar,
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	generateInfo(task)

	go log.InfoLogger.Println(log.Format(task, "Generating Session", "white"))

	// CHANGE
	task.FansaleVariables.EventCounter = 7478289
	task.FansaleVariables.LastCounter = 7478289
	task.FansaleVariables.Keywords = []string{
		// DE keywords
		"17348408",
		"17348446",
		"17348350",
		"17318939",
		"17333071",
		"17337297",
		"17337788",
		"17335909",

		// CH keywords
		"17337756",
		"17244455",

		// AT keywords
		"17275983",
		"17337851",
		"17337852",
	}

	for {
		for task.FansaleVariables.LastCounter+5 >= task.FansaleVariables.EventCounter {
			go log.InfoLogger.Println(log.Format(task, fmt.Sprintf("Checking %d", task.FansaleVariables.EventCounter), "white"))

			findOffers(task)
		}

		value := lastEntry()
		if value > 0 {
			task.FansaleVariables.LastCounter = value
		}

		task.FansaleVariables.EventCounter = task.FansaleVariables.LastCounter + 1
	}

	/*
		if task.Type == "Akamai" {
			for !session(task) {
				select {
				case <-task.Ctx.Done():
					return
				default:

					continue
				}
			}

			var counter int

			for {
				select {
				case <-task.Ctx.Done():
					return
				default:
					counter += 1

					getOffers(task)

					if counter%3 == 0 {
						// Submitting SensorData
						solveAkamai(task, rand.Intn(2)+2, task.FansaleVariables.AkamaiConfig, 300, 1300, task.FansaleVariables.Referer, true, false)
						time.Sleep(4 * time.Second)
					}

					continue
				}
			}
		} else if task.Type == "Seatmap" {
			for {
				select {
				case <-task.Ctx.Done():
					return
				default:
					seatmap(task)

					continue
				}
			}
		} else {
			for {
				select {
				case <-task.Ctx.Done():
					return
				default:
					detailSearch(task)

					continue
				}
			}
		}
	*/
}

func generateInfo(task *structs.Task) {
	task.FansaleVariables.UserAgent = utils.GetUserAgent(task)

	switch task.Site {
	case "www.fansale.de":
		task.FansaleVariables.Authority = "www.fansale.de"
		task.FansaleVariables.AffiliateId = "FAN"
	case "www.fansale.ch":
		task.FansaleVariables.Authority = "www.fansale.ch"
		task.FansaleVariables.AffiliateId = "FCH"
	case "www.fansale.it":
		task.FansaleVariables.Authority = "www.fansale.it"
		task.FansaleVariables.AffiliateId = "FIT"
	case "www.fansale.at":
		task.FansaleVariables.Authority = "www.fansale.at"
		task.FansaleVariables.AffiliateId = "FAU"
	case "www.fansale.pl":
		task.FansaleVariables.Authority = "www.fansale.pl"
		task.FansaleVariables.AffiliateId = "FPL"
	case "www.fansale.es":
		task.FansaleVariables.Authority = "www.fansale.es"
		task.FansaleVariables.AffiliateId = "FES"
	case "www.fansale.se":
		task.FansaleVariables.Authority = "www.fansale.se"
		task.FansaleVariables.AffiliateId = "FSE"
	case "www.fansale.no":
		task.FansaleVariables.Authority = "www.fansale.no"
		task.FansaleVariables.AffiliateId = "FNO"
	case "www.fansale.fi":
		task.FansaleVariables.Authority = "www.fansale.fi"
		task.FansaleVariables.AffiliateId = "FFI"
	default:
		task.FansaleVariables.Authority = "www.fansale.de"
		task.FansaleVariables.AffiliateId = "FAN"
	}

	task.Site = task.FansaleVariables.Authority

	task.FansaleVariables.EventId = strings.ReplaceAll(task.Input, " ", "")
	// Convert url to eventId
	if strings.Contains(task.FansaleVariables.EventId, "http") {
		// Detects all numbers from the url
		re := regexp.MustCompile(`[^\d]`)
		task.FansaleVariables.EventId = re.ReplaceAllString(strings.Split(task.FansaleVariables.EventId, "?")[0], "")
	}

	task.Event.Url = fmt.Sprintf("https://%s/fansale/searchresult/event/%s", task.FansaleVariables.Authority, task.FansaleVariables.EventId)
}

func findOffers(task *structs.Task) bool {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s/fansale/taylor-swift-the-eras-tour-tickets-%d.htm?bm-verify=1&affiliate=%s&_=%d", task.FansaleVariables.Authority, task.FansaleVariables.EventCounter, task.FansaleVariables.AffiliateId, time.Now().UnixMilli()), nil)

	req.Header = http.Header{
		"Connection":                {"keep-alive"},
		"Cache-Control":             {"no-cache"},
		"Pragma":                    {"no-cache"},
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
			"host", "connection", "cache-control", "pragma", "sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go log.ErrorLogger.Println(log.Format(task, "Proxy Error", "red"))
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)
		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		body, _, _ := utils.ParseResponse(res)

		for _, kw := range task.FansaleVariables.Keywords {
			if strings.Contains(body, kw) {
				// Check if there is an entry
				if checkExpiry(task.FansaleVariables.EventCounter) {
					go log.InfoLogger.Println(log.Format(task, "Detected Restock", "white"))

					notification.SendDiscord(task.Site+"_New", kw, fmt.Sprintf("https://%s/fansale/taylor-swift-the-eras-tour-tickets-%d.htm?affiliate=%s", task.FansaleVariables.Authority, task.FansaleVariables.EventCounter, task.FansaleVariables.AffiliateId), "")
				}

				break
			}
		}

		addEntry(task.FansaleVariables.EventCounter)

		task.FansaleVariables.LastCounter = task.FansaleVariables.EventCounter
		task.FansaleVariables.EventCounter += 1

		time.Sleep(time.Duration(task.Delay) * time.Millisecond)
		return false
	} else if res.StatusCode == 400 {
		go log.InfoLogger.Println(log.Format(task, "Waiting for new Offer", "white"))
	} else if res.StatusCode == 403 {
		if res.ContentLength == 0 {
			go log.InfoLogger.Println(log.Format(task, "Waiting for Offer", "white"))

			task.FansaleVariables.EventCounter += 1
		} else {
			go log.ErrorLogger.Println(log.Format(task, "Monitor Error - (403)", "red"))

			time.Sleep(time.Second)

			//task.Jar, _ = cookiejar.New(nil)
			//task.Client.Jar = task.Jar

			utils.ChangeRoundtripper(task, task.Client)
		}
	} else if res.StatusCode == 429 {
		go log.ErrorLogger.Println(log.Format(task, "Monitor Error - Rate Limit", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 503 {
		go log.InfoLogger.Println(log.Format(task, "Detected Queue", "magenta"))

		time.Sleep(30 * time.Second)
		return false
	} else {
		go log.ErrorLogger.Println(log.Format(task, fmt.Sprintf("Monitor Error - (%d)", res.StatusCode), "red"))

		task.FansaleVariables.EventCounter += 1
	}

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)
	return false
}

func getOffers(task *structs.Task) bool {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s/fansale/json/offers/%s?_=%d", task.FansaleVariables.Authority, task.FansaleVariables.EventId, rand.Intn(100000000000)+100000000000), nil)

	req.Header = http.Header{
		"Connection":                {"keep-alive"},
		"sec-ch-ua":                 {utils.GetSecChUa(task.FansaleVariables.UserAgent)},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"Windows"`},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {task.FansaleVariables.UserAgent},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Sec-Fetch-Site":            {"same-origin"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-User":            {"?1"},
		"Sec-Fetch-Dest":            {"document"},
		"Referer":                   {"https://" + task.FansaleVariables.Authority + "/fansale/"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"host", "connection", "sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "referer", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go log.ErrorLogger.Println(log.Format(task, "Proxy Error", "red"))
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)
		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		_, body, _ := utils.ParseResponse(res)

		var response seatmapResponse
		json.Unmarshal(body, &response)

		var tickets []string
		for _, ticket := range response.Seats {
			var skip bool
			for _, t := range tickets {
				if skip {
					break
				}

				if t != strconv.Itoa(ticket.ExtPlaceID) {
					for _, r := range ticket.RelatedExtPlaceIds {
						if strconv.Itoa(r) == t {
							skip = true
							break
						}
					}
				}
			}

			if skip {
				continue
			}

			tickets = append(tickets, strconv.Itoa(ticket.ExtPlaceID))
		}

		if len(tickets) > 0 && !reflect.DeepEqual(task.FansaleVariables.Tickets, tickets) {
			go log.InfoLogger.Println(log.Format(task, "Detected Restock", "white"))

			if task.FansaleVariables.RetryCounter > 0 {
				go notification.SendDiscord(task.Site+"_JSON", task.FansaleVariables.EventId, fmt.Sprintf("https://%s/fansale/searchresult/event/%s", task.FansaleVariables.Authority, task.FansaleVariables.EventId), "")

				for _, ticket := range tickets {
					var contains bool
					for _, t := range task.FansaleVariables.Tickets {
						if t == ticket {
							contains = true
							break
						}
					}

					if !contains {
						go log.InfoLogger.Println(log.Format(task, "Checking Seat - "+ticket, "white"))

						getOfferId(task, ticket)
						time.Sleep(time.Second)
					}
				}
			}

			task.FansaleVariables.Tickets = tickets

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		task.FansaleVariables.RetryCounter += 1

		go log.InfoLogger.Println(log.Format(task, "Waiting for Seats", "white"))
	} else if res.StatusCode == 403 {
		go log.ErrorLogger.Println(log.Format(task, "Monitor Error - (403)", "red"))

		task.Jar, _ = cookiejar.New(nil)
		task.Client.Jar = task.Jar

		time.Sleep(time.Duration(rand.Intn(5)+5) * time.Second)

		utils.ChangeRoundtripper(task, task.Client)

		for !session(task) {
			select {
			case <-task.Ctx.Done():
				return false
			default:
				continue
			}
		}
	} else if res.StatusCode == 429 {
		go log.ErrorLogger.Println(log.Format(task, "Monitor Error - Rate Limit", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 503 {
		go log.InfoLogger.Println(log.Format(task, "Detected Queue", "magenta"))

		time.Sleep(30 * time.Second)
		return false
	} else {
		go log.ErrorLogger.Println(log.Format(task, fmt.Sprintf("Monitor Error - (%d)", res.StatusCode), "red"))
	}

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)
	return false
}

type seatmapResponse struct {
	Seats []struct {
		ExtPlaceID         int   `json:"extPlaceId"`
		RelatedExtPlaceIds []int `json:"relatedExtPlaceIds"`
	} `json:"seats"`
	GeneralAdmissions []struct {
		ExtPlaceID         int   `json:"extPlaceId"`
		RelatedExtPlaceIds []any `json:"relatedExtPlaceIds"`
	} `json:"generalAdmissions"`
}

func seatmap(task *structs.Task) bool {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s/fansale/seatmap/tickets?eventId=%s&_=%d", task.FansaleVariables.Authority, task.FansaleVariables.EventId, rand.Intn(100000000000)+100000000000), nil)

	req.Header = http.Header{
		"Connection":                {"keep-alive"},
		"sec-ch-ua":                 {utils.GetSecChUa(task.FansaleVariables.UserAgent)},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"Windows"`},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {task.FansaleVariables.UserAgent},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Sec-Fetch-Site":            {"same-origin"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-User":            {"?1"},
		"Sec-Fetch-Dest":            {"document"},
		"Referer":                   {"https://" + task.FansaleVariables.Authority + "/fansale/"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"host", "connection", "sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "referer", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go log.ErrorLogger.Println(log.Format(task, "Proxy Error", "red"))
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)
		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		_, body, _ := utils.ParseResponse(res)

		var response seatmapResponse
		json.Unmarshal(body, &response)

		var tickets []string
		for _, ticket := range response.Seats {
			var skip bool
			for _, t := range tickets {
				if skip {
					break
				}

				if t != strconv.Itoa(ticket.ExtPlaceID) {
					for _, r := range ticket.RelatedExtPlaceIds {
						if strconv.Itoa(r) == t {
							skip = true
							break
						}
					}
				}
			}

			if skip {
				continue
			}

			tickets = append(tickets, strconv.Itoa(ticket.ExtPlaceID))
		}

		if len(tickets) > 0 && !reflect.DeepEqual(task.FansaleVariables.Tickets, tickets) {
			go log.InfoLogger.Println(log.Format(task, "Detected Restock", "white"))

			if task.FansaleVariables.RetryCounter > 0 {
				go notification.SendDiscord(task.Site+"_JSON", task.FansaleVariables.EventId, fmt.Sprintf("https://%s/fansale/searchresult/event/%s", task.FansaleVariables.Authority, task.FansaleVariables.EventId), "")

				for _, ticket := range tickets {
					var contains bool
					for _, t := range task.FansaleVariables.Tickets {
						if t == ticket {
							contains = true
							break
						}
					}

					if !contains {
						go log.InfoLogger.Println(log.Format(task, "Checking Seat - "+ticket, "white"))

						getOfferId(task, ticket)
						time.Sleep(time.Second)
					}
				}
			}

			task.FansaleVariables.Tickets = tickets

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		task.FansaleVariables.RetryCounter += 1

		go log.InfoLogger.Println(log.Format(task, "Waiting for Seats", "white"))
	} else if res.StatusCode == 403 {
		go log.ErrorLogger.Println(log.Format(task, "Monitor Error - (403)", "red"))

		task.Jar, _ = cookiejar.New(nil)
		task.Client.Jar = task.Jar

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 429 {
		go log.ErrorLogger.Println(log.Format(task, "Monitor Error - Rate Limit", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 503 {
		go log.InfoLogger.Println(log.Format(task, "Detected Queue", "magenta"))

		time.Sleep(30 * time.Second)
		return false
	} else {
		go log.ErrorLogger.Println(log.Format(task, fmt.Sprintf("Monitor Error - (%d)", res.StatusCode), "red"))
	}

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)
	return false
}

func detailSearch(task *structs.Task) bool {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s/fansale/detailSearch.htm?affiliate=%s&language=de&mobile=true&evId=%s&tab=false&_=%d", task.FansaleVariables.Authority, task.FansaleVariables.AffiliateId, task.FansaleVariables.EventId, rand.Intn(100000000000)+100000000000), nil)

	req.Header = http.Header{
		"Connection":                {"keep-alive"},
		"sec-ch-ua":                 {utils.GetSecChUa(task.FansaleVariables.UserAgent)},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"Windows"`},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {task.FansaleVariables.UserAgent},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Sec-Fetch-Site":            {"same-origin"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-User":            {"?1"},
		"Sec-Fetch-Dest":            {"document"},
		"Referer":                   {"https://" + task.FansaleVariables.Authority + "/fansale/"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"host", "connection", "sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "referer", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go log.ErrorLogger.Println(log.Format(task, "Proxy Error", "red"))
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)
		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		body, _, _ := utils.ParseResponse(res)

		doc, err := htmlquery.Parse(strings.NewReader(body))
		if err != nil {
			go log.ErrorLogger.Println(task, "Cannot Retrieve Information", "red")
			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			utils.ChangeRoundtripper(task, task.Client)
			return false
		}

		nodes, _ := htmlquery.QueryAll(doc, "//a[@class]")
		if len(nodes) == 0 {
			go log.ErrorLogger.Println(log.Format(task, "Waiting for Restock", "white"))

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		var tickets []string
		for _, node := range nodes {
			href := htmlquery.SelectAttr(node, "href")
			if !strings.Contains(href, "offerId") {
				continue
			}

			offerId := strings.Split(href, "offerId=")[1]

			tickets = append(tickets, offerId)
		}

		if len(tickets) > 0 && !reflect.DeepEqual(task.FansaleVariables.Tickets, tickets) {
			go log.InfoLogger.Println(log.Format(task, "Detected Restock", "white"))

			task.FansaleVariables.Tickets = tickets

			if task.FansaleVariables.RetryCounter > 0 {
				notification.SendDiscord(task.Site+"_HTML", task.FansaleVariables.EventId, fmt.Sprintf("https://%s/fansale/searchresult/event/%s", task.FansaleVariables.Authority, task.FansaleVariables.EventId), "")
			}

			time.Sleep(time.Duration(task.Delay) * time.Millisecond)
			return false
		}

		task.FansaleVariables.RetryCounter += 1

		go log.InfoLogger.Println(log.Format(task, "Waiting for Seats", "white"))
	} else if res.StatusCode == 403 {
		go log.ErrorLogger.Println(log.Format(task, "Monitor Error - (403)", "red"))

		task.Jar, _ = cookiejar.New(nil)
		task.Client.Jar = task.Jar

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 429 {
		go log.ErrorLogger.Println(log.Format(task, "Monitor Error - Rate Limit", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 503 {
		go log.InfoLogger.Println(log.Format(task, "Detected Queue", "magenta"))

		time.Sleep(30 * time.Second)
		return false
	} else {
		go log.ErrorLogger.Println(log.Format(task, fmt.Sprintf("Monitor Error - (%d)", res.StatusCode), "red"))
	}

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)
	return false
}

func getOfferId(task *structs.Task, seat string) bool {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s/fansale/seatmap/seats?eventId=%s&extPlaceId=%s&_=%d", task.FansaleVariables.Authority, task.FansaleVariables.EventId, seat, rand.Intn(100000000000)+100000000000), nil)

	req.Header = http.Header{
		"Connection":                {"keep-alive"},
		"sec-ch-ua":                 {utils.GetSecChUa(task.FansaleVariables.UserAgent)},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"Windows"`},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {task.FansaleVariables.UserAgent},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Sec-Fetch-Site":            {"same-origin"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-User":            {"?1"},
		"Sec-Fetch-Dest":            {"document"},
		"Referer":                   {"https://" + task.FansaleVariables.Authority + "/fansale/"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		http.HeaderOrderKey: {
			"host", "connection", "sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "referer", "accept-encoding", "accept-language",
		},
		http.PHeaderOrderKey: {
			":method", ":authority", ":scheme", ":path",
		},
	}

	res, err := task.Client.Do(req)
	if err != nil {
		go log.ErrorLogger.Println(log.Format(task, "Proxy Error", "red"))
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)
		utils.ChangeRoundtripper(task, task.Client)
		return false
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		body, _, _ := utils.ParseResponse(res)

		notification.SendDiscord(task.Site+"_JSON", task.FansaleVariables.EventId, fmt.Sprintf("https://%s/fansale/searchresult/event/%s", task.FansaleVariables.Authority, task.FansaleVariables.EventId), body)
	} else if res.StatusCode == 403 {
		go log.ErrorLogger.Println(log.Format(task, "Monitor Error - (403)", "red"))

		task.Jar, _ = cookiejar.New(nil)
		task.Client.Jar = task.Jar

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 429 {
		go log.ErrorLogger.Println(log.Format(task, "Monitor Error - Rate Limit", "red"))

		utils.ChangeRoundtripper(task, task.Client)
	} else if res.StatusCode == 503 {
		go log.InfoLogger.Println(log.Format(task, "Detected Queue", "magenta"))

		time.Sleep(30 * time.Second)
		return false
	} else {
		go log.ErrorLogger.Println(log.Format(task, fmt.Sprintf("Monitor Error - (%d)", res.StatusCode), "red"))
	}

	time.Sleep(time.Duration(task.Delay) * time.Millisecond)
	return false
}
