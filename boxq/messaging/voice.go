package messaging

type Voice struct {
	URL      string `json:"URl"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type"` // optional
	FileSize int    `json:"file_size"` // optional
}
