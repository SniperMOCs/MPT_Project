package content

type UpdatePostDTO struct {
	Description *string  `json:"description"`
	IsVisible   *bool    `json:"is_visible"`
	MediaURLs   []string `json:"media_urls"`
}
