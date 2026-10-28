package model

type Link struct {
	Code        string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
