package content

type ContentDTO struct {
	ContentType ContentType `json:"content_type"`
	Description string      `json:"description"`
}
