package structs

import (
	"context"
	http "github.com/bogdanfinn/fhttp"
	"sync"
)

type Task struct {
	Mutex     sync.Mutex         `json:"-"`
	Mode      string             `json:"-"`
	Ctx       context.Context    `json:"-"`
	Cancel    context.CancelFunc `json:"-"`
	Code      string             `json:"-"`
	Active    bool               `json:"-"`
	ProxyFile string             `json:"-"`
	Site      string             `json:"-"`
	Id        string             `json:"-"`
	Type      string             `json:"-"`
	GroupId   string             `json:"-"`
	Input     string             `json:"-"`
	Delay     int                `json:"-"`
	Event     Event              `json:"-"`
	Client    *http.Client       `json:"-"`
	Jar       http.CookieJar     `json:"-"`
	Proxy     string             `json:"-"`
	Security  struct {
		Jwt string
	} `json:"-"`
	EventimVariables      eventimUtils      `json:"-"`
	FansaleVariables      fansaleUtils      `json:"-"`
	TicketmasterVariables ticketmasterUtils `json:"-"`
}

type Event struct {
	Name     string `json:"name"`
	EventId  string `json:"eventId"`
	Venue    string `json:"venue"`
	Date     string `json:"date"`
	IsResale string `json:"isResale"`
	Row      string `json:"row"`
	Section  string `json:"section"`
	Seat     string `json:"seat"`
	Price    string `json:"price"`
	Url      string `json:"url"`
	Image    string `json:"image"`
	Expiry   string `json:"expiry"`
}
