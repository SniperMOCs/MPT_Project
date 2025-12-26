package content

// UpdatePostDTO представляет данные для обновления поста
// @Description Данные для обновления существующего поста
type UpdatePostDTO struct {
	Description *string  `json:"description,omitempty" example:"Updated description"`                // Описание поста (опционально)
	IsVisible   *bool    `json:"is_visible,omitempty" example:"true"`                                // Видимость поста (опционально)
	MediaURLs   []string `json:"media_urls,omitempty" example:"[/uploads/content/post1/image1.jpg]"` // Массив URL медиа-файлов (опционально)
}
