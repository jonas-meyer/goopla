package goopla

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type ListingService struct {
	client *Client
}

type ListingResponse struct {
	SearchSettings
	Listings []Listing `xml:"listing"`
}

type SearchSettings struct {
	Area      string `xml:"area_name"`
	Street    string `xml:"street"`
	Town      string `xml:"town"`
	County    string `xml:"county"`
	Country   string `xml:"country"`
	PostCode  string `xml:"postcode"`
	Latitude  string `xml:"latitude"`
	Longitude string `xml:"longitude"`
	SearchBoundingBox
	ListingAmount int `xml:"result_count"`
}

type SearchBoundingBox struct {
	MaxLatitude  string `xml:"bounding_box>latitude_max"`
	MinLatitude  string `xml:"bounding_box>latitude_min"`
	Maxlongitude string `xml:"bounding_box>longitude_max"`
	Minlongitude string `xml:"bounding_box>longitude_min"`
}

type Listing struct {
	Agent
	ListingDetails
	RoomDetails
	RentalPrices
}

type Agent struct {
	AgentName     string `xml:"agent_name"`
	AgentID       string `xml:"agent_id"`
	AgentAddress  string `xml:"company_address"`
	AgentLogo     string `xml:"agent_logo"`
	AgentPhone    string `xml:"agent_phone"`
	AgentCategory string `xml:"category"`
}

type ListingDetails struct {
	ListingID        string  `xml:"listing_id"`
	ListingURL       XMLURL  `xml:"details_url"`
	ImageURL         XMLURL  `xml:"image_url"`
	Address          string  `xml:"displayable_address"`
	Town             string  `xml:"post_town"`
	PostCode         string  `xml:"outcode"`
	AvailabilityDate XMLDate `xml:"available_from_display"`
	FirstPublished   XMLTime `xml:"first_published_date"`
	LastPublished    XMLTime `xml:"last_published_date"`
	Status           string  `xml:"status"`
	Description      string  `xml:"description"`
	ShortDescription string  `xml:"short_description"`
	PropertyType     string  `xml:"property_type"`
	FloorPlan        XMLURL  `xml:"floor_plan"`
	FurnishedState   string  `xml:"furnished_state"`
	LettingFees      string  `xml:"letting_fees"`
	SharedOccupancy  string  `xml:"rental_prices>shared_occupancy"`
}

type RoomDetails struct {
	Bedrooms  int `xml:"num_bedrooms"`
	Bathrooms int `xml:"num_bathrooms"`
	Floors    int `xml:"num_floors"`
	Recepts   int `xml:"num_recepts"`
	FloorArea `xml:"floor_area"`
}

type RentalPrices struct {
	AccurateAmount string `xml:"rental_prices>accurate"`
	Monthly        int    `xml:"rental_prices>per_month"`
	Weekly         int    `xml:"rental_prices>per_week"`
}

type FloorArea struct {
	Name  string `xml:"name"`
	Units string `xml:"units"`
	Value string `xml:"value"`
}

type XMLTime time.Time
type XMLURL url.URL
type XMLDate time.Time

func (x *XMLTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}
	*x = decodeDate(content)
	return nil
}

func (x *XMLDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}
	*x = decodeAvailabilityDate(content)
	return nil
}

func decodeDate(s string) XMLTime {
	dateLayout := "2006-01-02 15:04:05"
	t, _ := time.Parse(dateLayout, s)
	return XMLTime(t)
}

func decodeAvailabilityDate(s string) XMLDate {
	if s == "Available immediately" {
		return XMLDate(time.Now())
	}
	dateLayout := "2 Jan 2006"
	re := regexp.MustCompile(`\d+(?:st|nd|rd|th) \w+ \d{4}`)
	dateString := re.FindString(s)
	dateStringSuffixRemoved := regexp.MustCompile(`(?P<day>\d+)(?P<suffix>st|nd|rd|th)`).ReplaceAllString(dateString, "${day}")
	date, _ := time.Parse(dateLayout, dateStringSuffixRemoved)
	return XMLDate(date)
}

func (x *XMLDate) UnmarshalXMLAttr(attr xml.Attr) error {
	*x = decodeAvailabilityDate(attr.Value)
	return nil
}

func (x *XMLTime) UnmarshalXMLAttr(attr xml.Attr) error {
	*x = decodeDate(attr.Value)
	return nil
}

func (u *XMLURL) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}
	*u = decodeURL(content)
	return nil
}

func decodeURL(s string) XMLURL {
	url, _ := url.Parse(s)
	return XMLURL(*url)
}

func (u *XMLURL) UnmarshalXMLAttr(attr xml.Attr) error {
	*u = decodeURL(attr.Value)
	return nil
}

func (d *ListingDetails) MarshalJSON() ([]byte, error) {
	type Alias ListingDetails
	return json.Marshal(&struct {
		ListingID        string `json:"ListingID"`
		FirstPublished   string `json:"FirstPublished"`
		LastPublished    string `json:"LastPublished"`
		AvailabilityDate string `json:"AvailabilityDate"`
		*Alias
	}{
		ListingID:        d.ListingID,
		FirstPublished:   time.Time(d.FirstPublished).Format(time.RFC3339),
		LastPublished:    time.Time(d.LastPublished).Format(time.RFC3339),
		AvailabilityDate: time.Time(d.AvailabilityDate).Format(time.RFC3339),
		Alias:            (*Alias)(d),
	})
}

func (d *ListingDetails) UnmarshalJSON(data []byte) error {
	type Alias ListingDetails
	aux := &struct {
		ListingID        string `json:"ListingID"`
		FirstPublished   string `json:"FirstPublished"`
		LastPublished    string `json:"LastPublished"`
		AvailabilityDate string `json:"AvailabilityDate"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	d.ListingID = aux.ListingID
	firstPublished, _ := time.Parse(time.RFC3339, aux.FirstPublished)
	LastPublished, _ := time.Parse(time.RFC3339, aux.LastPublished)
	AvailabilityDate, _ := time.Parse(time.RFC3339, aux.AvailabilityDate)
	d.FirstPublished = XMLTime(firstPublished)
	d.LastPublished = XMLTime(LastPublished)
	d.AvailabilityDate = XMLDate(AvailabilityDate)
	return nil
}

func (l *ListingService) Get(ctx context.Context, opts *ListingOptions) (*ListingResponse, *http.Response, error) {
	path := fmt.Sprintf("property_listings.xml?api_key=%s", l.client.Credentials.ApiKey)

	verifiedOptions, err := verifyOptions(opts)
	if err != nil {
		return nil, nil, err
	}

	path, err = addOptions(path, verifiedOptions)
	if err != nil {
		return nil, nil, err
	}

	req, err := l.client.NewRequest(path, url.Values{})
	if err != nil {
		return nil, nil, err
	}

	listings := new(ListingResponse)
	resp, err := l.client.Do(ctx, req, listings)

	if err != nil {
		return nil, resp, err
	}

	return listings, resp, nil
}

func verifyOptions(opts *ListingOptions) (*ListingOptions, error) {
	return opts, nil
}
