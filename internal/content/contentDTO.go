package content

// ContentDTO представляет данные для создания поста
// @Description Данные для создания нового поста
type ContentDTO struct {
	Description string `json:"description" example:"Beautiful sunset!"` // Описание поста
}
