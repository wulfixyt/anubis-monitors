package structs

type eventimUtils struct {
	AkamaiUrl string
	Authority string
	UserAgent struct {
		App string
		Web string
	}
	HandsOnDTO    interface{}
	Categoryies   []string
	ReloadUrl     string
	Referer       string
	Path          string
	SessionStart  int64
	Token         string
	AffiliateId   string
	RequiresPromo bool
	PromoCode     string
}

type fansaleUtils struct {
	Authority    string
	AffiliateId  string
	AkamaiUrl    string
	AkamaiConfig string
	EventId      string
	EventCounter int
	LastCounter  int
	Referer      string
	RetryCounter int
	Tickets      []string
	UserAgent    string
	Keywords     []string
}

type ticketmasterUtils struct {
	Authority string
	UserAgent string
}
