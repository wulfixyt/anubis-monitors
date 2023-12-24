package eventim

type priceInfo struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Reductions []struct {
		ID                  string        `json:"id"`
		Name                string        `json:"name"`
		Holds               []interface{} `json:"holds"`
		Quantities          []int         `json:"quantities"`
		PriceText           string        `json:"priceText"`
		TdlRabattLinkId     int           `json:"tdlRabattLinkId"`
		Available           bool          `json:"available"`
		DefaultReduction    bool          `json:"defaultReduction"`
		PromotionReduction  bool          `json:"promotionReduction"`
		Price               float64       `json:"price"`
		Currency            string        `json:"currency"`
		MemberclubExclusive bool          `json:"memberclubExclusive"`
		MarketingLabels     []interface{} `json:"marketingLabels"`
		MemberclubText      interface{}   `json:"memberclubText"`
		HasCttRules         bool          `json:"hasCttRules"`
	} `json:"reductions"`
	Gtr                interface{}   `json:"gtr"`
	PriceText          interface{}   `json:"priceText"`
	Available          bool          `json:"available"`
	BackGroundColor    []int         `json:"backGroundColor"`
	TextColor          []int         `json:"textColor"`
	Category           string        `json:"category"`
	MarketingLabels    []interface{} `json:"marketingLabels"`
	ApplicableCttRules interface{}   `json:"applicableCttRules"`
	MaxAvailability    int           `json:"maxAvailability"`
}

type handsOnDTO struct {
	Version           int `json:"version"`
	PromotionTypeInfo []struct {
		Closed bool   `json:"closed"`
		Id     int    `json:"id"`
		Title  string `json:"title"`
	} `json:"promotionTypeInfo"`
	SelectedPromotionInfo interface{}   `json:"selectedPromotionInfo"`
	EventSeriesId         int           `json:"eventSeriesId"`
	PriceInfo             []priceInfo   `json:"priceInfo"`
	CttRules              []interface{} `json:"cttRules"`
	MaxAvailability       int           `json:"maxAvailability"`
	ActiveCttRules        interface{}   `json:"activeCttRules"`
	PromotionMandatory    bool          `json:"promotionMandatory"`
	NewActiveCttRules     interface{}   `json:"newActiveCttRules"`
	PromotionCode         interface{}   `json:"promotionCode"`
	EventId               int           `json:"eventId"`
	MarketingLabels       []interface{} `json:"marketingLabels"`
	SelectedPromotion     int           `json:"selectedPromotion"`
	RequestedPromotion    int           `json:"requestedPromotion"`
	ResetPromotionId      interface{}   `json:"resetPromotionId"`
	SelectionMode         string        `json:"selectionMode"`
	Groups                []groups      `json:"groups"`
	ExclusivePromotion    bool          `json:"exclusivePromotion"`
}

type groups struct {
	Id    string  `json:"id"`
	Seats []seats `json:"seats"`
}

type seats struct {
	PriceCategoryID string `json:"priceCategoryId"`
}

type codePayload struct {
	HandsOnDTO handsOnDTO `json:"handsOnDTO"`
}

type codeResponse struct {
	Status struct {
		Code          int         `json:"code"`
		IsWaitingroom interface{} `json:"isWaitingroom"`
		HttpCode      int         `json:"httpCode"`
		Data          struct {
			Routing []struct {
				RoutingURL struct {
					UrlType string      `json:"urlType"`
					Url     string      `json:"url"`
					Params  interface{} `json:"params"`
				} `json:"routingUrl"`
			} `json:"routing"`
			Success bool `json:"success"`
		} `json:"data"`
	} `json:"status"`
}

type monitorResponse struct {
	Status struct {
		Code          int         `json:"code"`
		Text          string      `json:"text"`
		IsWaitingRoom interface{} `json:"isWaitingRoom"`
		HttpCode      int         `json:"httpCode"`
	} `json:"status"`
	Section struct {
		Model struct {
			EventSeriesId int        `json:"eventSeriesId"`
			TicketTitle   string     `json:"ticketTitle"`
			EventLocation string     `json:"eventLocation"`
			HandsOnDTO    handsOnDTO `json:"handsOnDTO"`
		} `json:"model"`
	} `json:"section"`
}