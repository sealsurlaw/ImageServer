package response

import "time"

type UploadImageResponse struct {
	Filename string `json:"filename"`
}

type GetLinkResponse struct {
	Url       string     `json:"url"`
	ExpiresAt *time.Time `json:"expiresAt"`
}
