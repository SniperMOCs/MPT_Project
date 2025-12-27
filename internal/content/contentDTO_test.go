package content

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestContentDTO_JSON(t *testing.T) {
	t.Run("Сериализация ContentDTO", func(t *testing.T) {
		dto := ContentDTO{
			Description: "Test description",
		}

		data, err := json.Marshal(dto)
		if err != nil {
			t.Errorf("Ошибка при сериализации ContentDTO: %v", err)
		}

		var decoded ContentDTO
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Ошибка при десериализации ContentDTO: %v", err)
		}

		if decoded.Description != dto.Description {
			t.Errorf("Description = %s, ожидался %s", decoded.Description, dto.Description)
		}
	})

	t.Run("Сериализация UpdatePostDTO", func(t *testing.T) {
		description := "Updated description"
		isVisible := true
		mediaURLs := []string{"/uploads/image1.jpg", "/uploads/image2.jpg"}

		dto := UpdatePostDTO{
			Description: &description,
			IsVisible:   &isVisible,
			MediaURLs:   mediaURLs,
		}

		data, err := json.Marshal(dto)
		if err != nil {
			t.Errorf("Ошибка при сериализации UpdatePostDTO: %v", err)
		}

		var decoded UpdatePostDTO
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Ошибка при десериализации UpdatePostDTO: %v", err)
		}

		if *decoded.Description != description {
			t.Errorf("Description = %s, ожидался %s", *decoded.Description, description)
		}
		if *decoded.IsVisible != isVisible {
			t.Errorf("IsVisible = %v, ожидался %v", *decoded.IsVisible, isVisible)
		}
		if len(decoded.MediaURLs) != 2 {
			t.Errorf("MediaURLs содержит %d элементов, ожидалось 2", len(decoded.MediaURLs))
		}
	})

	t.Run("Сериализация UpdatePostDTO с omitempty", func(t *testing.T) {
		dto := UpdatePostDTO{}

		data, err := json.Marshal(dto)
		if err != nil {
			t.Errorf("Ошибка при сериализации UpdatePostDTO: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("Ошибка при десериализации JSON: %v", err)
		}

		// Проверяем что поля с omitempty отсутствуют
		if _, ok := result["description"]; ok {
			t.Error("Поле description не должно присутствовать в JSON")
		}
		if _, ok := result["is_visible"]; ok {
			t.Error("Поле is_visible не должно присутствовать в JSON")
		}
		// media_urls не имеет omitempty, но пустой массив должен быть в JSON
		if _, ok := result["media_urls"]; !ok {
			t.Error("Поле media_urls должно присутствовать в JSON даже если пустое")
		}
	})
}

func TestContentModel_JSON(t *testing.T) {
	t.Run("Сериализация Content", func(t *testing.T) {
		content := Content{
			ID:          uuid.New(),
			AuthorId:    uuid.New(),
			Description: "Test content",
			MediaURLs:   []string{"/uploads/image1.jpg"},
			CreatedAt:   time.Now(),
			IsVisible:   true,
		}

		data, err := json.Marshal(content)
		if err != nil {
			t.Errorf("Ошибка при сериализации Content: %v", err)
		}

		var decoded Content
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Ошибка при десериализации Content: %v", err)
		}

		if decoded.Description != content.Description {
			t.Errorf("Description = %s, ожидался %s", decoded.Description, content.Description)
		}
		if decoded.IsVisible != content.IsVisible {
			t.Errorf("IsVisible = %v, ожидался %v", decoded.IsVisible, content.IsVisible)
		}
		if len(decoded.MediaURLs) != 1 {
			t.Errorf("MediaURLs содержит %d элементов, ожидалось 1", len(decoded.MediaURLs))
		}
	})
}

func TestContentResponse_JSON(t *testing.T) {
	t.Run("Сериализация ContentResponse", func(t *testing.T) {
		avatarURL := "/uploads/avatar.jpg"
		response := ContentResponse{
			ID:          uuid.New(),
			Description: "Test response",
			MediaURLs:   []string{"/uploads/image1.jpg"},
			CreatedAt:   time.Now(),
			IsVisible:   true,
			Author: ContentAuthor{
				ID:        uuid.New(),
				Username:  "testuser",
				AvatarURL: &avatarURL,
			},
		}

		data, err := json.Marshal(response)
		if err != nil {
			t.Errorf("Ошибка при сериализации ContentResponse: %v", err)
		}

		var decoded ContentResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Ошибка при десериализации ContentResponse: %v", err)
		}

		if decoded.Description != response.Description {
			t.Errorf("Description = %s, ожидался %s", decoded.Description, response.Description)
		}
		if decoded.Author.Username != response.Author.Username {
			t.Errorf("Author.Username = %s, ожидался %s", decoded.Author.Username, response.Author.Username)
		}
		if *decoded.Author.AvatarURL != *response.Author.AvatarURL {
			t.Errorf("Author.AvatarURL = %s, ожидался %s", *decoded.Author.AvatarURL, *response.Author.AvatarURL)
		}
	})

	t.Run("Сериализация ContentResponse с nil AvatarURL", func(t *testing.T) {
		response := ContentResponse{
			ID:          uuid.New(),
			Description: "Test response",
			MediaURLs:   []string{},
			CreatedAt:   time.Now(),
			IsVisible:   true,
			Author: ContentAuthor{
				ID:       uuid.New(),
				Username: "testuser",
			},
		}

		data, err := json.Marshal(response)
		if err != nil {
			t.Errorf("Ошибка при сериализации ContentResponse с nil AvatarURL: %v", err)
		}

		var decoded ContentResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Ошибка при десериализации ContentResponse с nil AvatarURL: %v", err)
		}

		if decoded.Author.AvatarURL != nil {
			t.Error("AvatarURL должен быть nil")
		}
	})
}
