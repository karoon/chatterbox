package messaging

type Video struct {
	FileID    string `json:"file_id"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Duration  int    `json:"duration"`
	Thumbnail *Photo `json:"thumb"`     // optional
	MimeType  string `json:"mime_type"` // optional
	FileSize  int    `json:"file_size"` // optional
}
