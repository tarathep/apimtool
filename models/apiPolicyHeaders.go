package models

import "encoding/xml"

type Policies struct {
	XMLName xml.Name `xml:"policies"`
	Text    string   `xml:",chardata"`
	Inbound struct {
		Text              string `xml:",chardata"`
		Base              string `xml:"base"`
		SetBackendService struct {
			Text      string `xml:",chardata"`
			BackendID string `xml:"backend-id,attr"`
		} `xml:"set-backend-service"`
		SetHeader []struct {
			Text         string `xml:",chardata"`
			Name         string `xml:"name,attr"`
			ExistsAction string `xml:"exists-action,attr"`
			Value        string `xml:"value"`
		} `xml:"set-header"`
	} `xml:"inbound"`
	Backend struct {
		Text string `xml:",chardata"`
		Base string `xml:"base"`
	} `xml:"backend"`
	Outbound struct {
		Text string `xml:",chardata"`
		Base string `xml:"base"`
	} `xml:"outbound"`
	OnError struct {
		Text string `xml:",chardata"`
		Base string `xml:"base"`
	} `xml:"on-error"`
}
