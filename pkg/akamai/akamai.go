package akamai

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
	"github.com/wulfixyt/anubis-monitors/pkg/utils"
)

func GetUserAgent(task *structs.Task) string {
	req, _ := http.NewRequest("GET", "https://ak01-eu.hwkapi.com/akamai/ua", nil)

	req.Header = http.Header{
		"Accept-Encoding": {"gzip, deflate"},
		"X-Api-Key":       {"69d6ae60-c7da-4c8d-b49f-0c5917409252"},
		"X-Sec":           {"new"},
	}

	resp, err := task.Client.Do(req)
	if err != nil {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	}

	defer resp.Body.Close()

	body, _, _ := utils.ParseResponse(resp)
	return body
}

func GetConfig(task *structs.Task, data []byte) string {
	hash := md5.Sum(data)
	req, _ := http.NewRequest("POST", "https://ak-ppsaua.hwkapi.com/006180d12cf7/c", strings.NewReader(fmt.Sprintf(`{"hash":"%s"}`, hex.EncodeToString(hash[:]))))

	req.Header = http.Header{
		"Accept-Encoding": {"gzip, deflate"},
		"Content-Type":    {"application/json"},
		"X-Api-Key":       {"69d6ae60-c7da-4c8d-b49f-0c5917409252"},
		"X-Sec":           {"new"},
	}

	resp, err := task.Client.Do(req)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	body, _, _ := utils.ParseResponse(resp)

	if body == "false" {
		body = func() string {
			req, _ := http.NewRequest("POST", "https://ak-ppsaua.hwkapi.com/006180d12cf7", strings.NewReader(fmt.Sprintf(`{"body":"%s"}`, base64.StdEncoding.EncodeToString(data))))

			req.Header = http.Header{
				"Accept-Encoding": {"gzip, deflate"},
				"Content-Type":    {"application/json"},
				"X-Api-Key":       {"69d6ae60-c7da-4c8d-b49f-0c5917409252"},
				"X-Sec":           {"new"},
			}

			resp, err := task.Client.Do(req)
			if err != nil {
				return ""
			}

			defer resp.Body.Close()

			body, _, _ = utils.ParseResponse(resp)
			return body
		}()
	}
	return body
}

func GetSensor(task *structs.Task, site string, abck string, bm_sz string, userAgent string, config string, mouseEvents bool, keyboardEvents bool) string {
	payload := map[string]string{}

	mouse := "0"
	if mouseEvents {
		mouse = "1"
	}

	keyboard := "0"
	if keyboardEvents {
		keyboard = "1"
	}

	payload["site"] = site
	payload["abck"] = abck
	payload["bm_sz"] = bm_sz
	payload["events"] = fmt.Sprintf("%s,%s", mouse, keyboard)
	payload["user_agent"] = userAgent

	if config != "" {
		payload["config"] = config
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}

	req, _ := http.NewRequest("POST", "https://ak01-eu.hwkapi.com/akamai/generate", bytes.NewBuffer(data))

	req.Header = http.Header{
		"Accept-Encoding": {"gzip, deflate"},
		"Content-Type":    {"application/json"},
		"X-Api-Key":       {"69d6ae60-c7da-4c8d-b49f-0c5917409252"},
		"X-Sec":           {"new"},
	}

	resp, err := task.Client.Do(req)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()
	body, _, _ := utils.ParseResponse(resp)
	if strings.Contains(body, "****") {
		return strings.Split(body, "****")[0]
	}

	return body
}
