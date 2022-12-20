package models

type API struct {
	Apiname  string   `json:"apiname"`
	Env      string   `json:"env"`
	Tags     []string `json:"tags"`
	Policies struct {
		BackendURL string `json:"backend-url"`
		SetHeaders []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"set-headers"`
	} `json:"policies"`
	Operations []struct {
		Name   string `json:"name"`
		Method string `json:"method"`
		URL    string `json:"url"`
	} `json:"operations"`
}
