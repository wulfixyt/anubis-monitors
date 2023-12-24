package utils

import (
	"bufio"
	"bytes"
	"client"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"

	"github.com/andybalholm/brotli"
)

var (
	legacyGreasyChars = []string{" ", " ", ";"}
	greasyChars       = []string{" ", "(", ":", "-", ".", "/", ")", ";", "=", "?", "_"}
	greasyVersion     = []string{"8", "99", "24"}
	greasyOrders      = [][]int{
		{0, 1, 2}, {0, 2, 1}, {1, 0, 2},
		{1, 2, 0}, {2, 0, 1}, {2, 1, 0},
	}

	BrandChrome Brand = "Google Chrome"
)

type Brand string

func ChangeRoundtripper(task *structs.Task, reqClient *http.Client) {
	reqClient.CloseIdleConnections()
	RotateProxy(task)

	rt, err := client.NewRoundtripper(profiles.Chrome_117, client.Settings{Proxy: task.Proxy})
	if err != nil {
		return
	}

	reqClient.Transport = rt
}

func RotateProxy(task *structs.Task) {
	file, err := os.Open("./proxies.txt")
	if err != nil {
		task.Proxy = ""
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var proxies []string
	for scanner.Scan() {
		proxies = append(proxies, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		task.Proxy = ""
		return
	}

	if len(proxies) > 0 {
		task.Proxy = splitProxy(proxies[rand.Intn(len(proxies))])
		return
	}

	task.Proxy = ""
	return
}

func splitProxy(input string) string {
	split := strings.Split(input, ":")
	if len(split) == 2 {
		proxy := "http://" + split[0] + ":" + split[1]
		return proxy
	} else if len(split) == 4 {
		proxy := "http://" + split[2] + ":" + split[3] + "@" + split[0] + ":" + split[1]
		return proxy
	}

	return ""
}

func ShortenUrl(task *structs.Task, longUrl string) string {
	resp, err := http.Post("https://api.anubisio.com/url/create", "application/json", bytes.NewBuffer([]byte(`{"url":"`+longUrl+`"}`)))
	if err != nil {
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)

		return ShortenUrl(task, longUrl)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)

		return ShortenUrl(task, longUrl)
	}

	type Response struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
	}

	var parsed Response
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)

		return ShortenUrl(task, longUrl)
	}

	if parsed.Success != true {
		time.Sleep(time.Duration(task.Delay) * time.Millisecond)

		return ShortenUrl(task, longUrl)
	}

	return parsed.Message
}

func ParseResponse(response *http.Response) (string, []byte, error) {
	body, _ := ioutil.ReadAll(response.Body)

	responseBody := string(body)
	encodeType := response.Header.Get("Content-Encoding")
	var buf2 bytes.Buffer

	if encodeType == "gzip" && responseBody != "" {
		err := gunzipWrite(&buf2, body)
		if err != nil {
			return responseBody, body, err
		}
		body = buf2.Bytes()

		responseBody = buf2.String()
	} else if encodeType == "br" && responseBody != "" {
		r := brotli.NewReader(bytes.NewReader(body))
		reader, err := ioutil.ReadAll(r)
		if err != nil {
			return responseBody, body, err
		}

		body = reader
		responseBody = string(reader)
	}

	return responseBody, body, nil
}

func gunzipWrite(w io.Writer, data []byte) error {
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer gr.Close()
	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}

// SendDebug - Debug
func SendDebug(task *structs.Task, responseUrl string, responseStatus string, responseHeaders http.Header, responseBody string) {
	type Request struct {
		Url     string `json:"url"`
		Headers string `json:"headers"`
		Body    string `json:"body"`
		Status  string `json:"status"`
	}

	p := Request{
		Url:     responseUrl,
		Status:  responseStatus,
		Headers: fmt.Sprintf("%v", responseHeaders),
		Body:    responseBody,
	}

	payload, _ := json.Marshal(p)

	resp, err := http.Post(fmt.Sprintf("http://logs.anubisio.com/%s/log?token=", strings.ReplaceAll(strings.ToLower(task.Site), " ", ""))+task.Security.Jwt, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	resp.Body.Close()
}

func GetValue(task *structs.Task, name string, domain string) string {
	cookieUrl, _ := url.Parse("https://" + domain)
	for _, cookie := range task.Client.Jar.Cookies(cookieUrl) {
		if cookie.Name == name {
			return cookie.Value
		}
	}

	return ""
}

func GetUserAgent(task *structs.Task) string {
	req, _ := http.NewRequest("GET", "https://ak01-eu.hwkapi.com/akamai/ua", nil)

	req.Header = http.Header{
		"Accept-Encoding": {"gzip, deflate"},
		"X-Api-Key":       {"69d6ae60-c7da-4c8d-b49f-0c5917409252"},
		"X-Sec":           {"new"},
	}

	resp, err := task.Client.Do(req)
	if err != nil {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
	}

	defer resp.Body.Close()

	body, _, _ := ParseResponse(resp)

	if !strings.Contains(body, "Chrome") {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
	}
	return body
}

func formatBrand(brand Brand, majorVersion string) string {
	return fmt.Sprintf(`"%s";v="%s"`, brand, majorVersion)
}

func greasedBrand(seed int, majorVersionNumber int, permutedOrder []int) string {
	var brand, version string

	switch {
	case majorVersionNumber <= 102, majorVersionNumber == 104:
		brand = fmt.Sprintf("%sNot%sA%sBrand", legacyGreasyChars[permutedOrder[0]], legacyGreasyChars[permutedOrder[1]], legacyGreasyChars[permutedOrder[2]])
		version = "99"
	case majorVersionNumber == 103:
		brand = fmt.Sprintf("%sNot%sA%sBrand", greasyChars[(seed%(len(greasyChars)-1))+1], greasyChars[(seed+1)%len(greasyChars)], greasyChars[(seed+2)%len(greasyChars)])
		version = greasyVersion[seed%len(greasyVersion)]
	default: // >=105
		// https://github.com/WICG/ua-client-hints/pull/310
		brand = fmt.Sprintf("Not%sA%sBrand", greasyChars[seed%len(greasyChars)], greasyChars[(seed+1)%len(greasyChars)])
		version = greasyVersion[seed%len(greasyVersion)]
	}

	return formatBrand(Brand(brand), version)
}

func getUaVersion(ua string) int {
	// assuming ua is a valid Chrome UA

	rawVersion := strings.Split(strings.Split(ua, "Chrome/")[1], ".")[0]
	version, err := strconv.Atoi(rawVersion)
	if err != nil {
		return 115
	}
	return version
}

func GetSecChUa(ua string) string {
	majorVersionNumber := getUaVersion(ua)
	majorVersion := strconv.Itoa(majorVersionNumber)
	seed := getUaVersion(ua)
	if seed <= 102 {
		// legacy behavior (maybe a bug?)
		seed = 0
	}

	order := greasyOrders[seed%len(greasyOrders)]

	greased := make([]string, 3)

	greased[order[0]] = greasedBrand(seed, majorVersionNumber, order)
	greased[order[1]] = formatBrand("Chromium", majorVersion)
	greased[order[2]] = formatBrand(BrandChrome, majorVersion)

	return strings.Join(greased, ", ")
}

func EncodeCookies(jar http.CookieJar, domain string) string {
	cookieMap := []map[string]string{}

	for _, cookie := range jar.Cookies(&url.URL{Scheme: "https", Host: domain, Path: "/"}) {
		cookieMap = append(cookieMap, map[string]string{"name": cookie.Name, "value": cookie.Value, "domain": cookie.Domain, "path": "/", "url": "https://" + domain + "/"})
	}

	rawCookies, _ := json.Marshal(cookieMap)

	cookieString := base64.StdEncoding.EncodeToString(rawCookies)

	return cookieString
}

func ImagetoText(task *structs.Task, b64image string) string {
	type Response struct {
		Message string `json:"message"`
	}

	payload := strings.NewReader(fmt.Sprintf(`{"data":"%s"}`, b64image))

	resp, err := http.Post("https://server.anubisio.com/V3VsZml4/ocr?token="+task.Security.Jwt, "application/json", payload)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	_, body, _ := ParseResponse(resp)

	var parsed Response
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return err.Error()
	}

	return parsed.Message
}

func SetCookies(task *structs.Task, jar []*http.Cookie, domain string) {
	for _, cookie := range jar {
		task.Client.Jar.SetCookies(&url.URL{Scheme: "https", Host: domain, Path: "/"}, []*http.Cookie{
			{
				Name:   cookie.Name,
				Value:  cookie.Value,
				Domain: domain,
			},
		})
	}
}

func DelCookie(client *http.Client, domain string, name string) {
	u, _ := url.Parse("https://" + domain)

	for _, cookie := range client.Jar.Cookies(u) {
		if strings.Contains(cookie.Name, name) || cookie.Name == name {
			cookie.MaxAge = -1
			client.Jar.SetCookies(u, []*http.Cookie{cookie})
		}
	}
}
