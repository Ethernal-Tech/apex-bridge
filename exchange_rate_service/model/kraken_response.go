package model

type KrakenResponse struct {
	Result map[string]struct {
		C []string `json:"c"`
	} `json:"result"`
}
