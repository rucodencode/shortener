package model

type Link struct {
	UserID      string `json:"user_id"`
	Code        string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
