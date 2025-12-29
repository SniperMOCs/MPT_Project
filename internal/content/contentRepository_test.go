package content

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentRepository_Create(t *testing.T) {
	t.Run("Успешное создание контента", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		content := &Content{
			ID:          uuid.New(),
			AuthorId:    uuid.New(),
			Description: "Test description",
			MediaURLs:   []string{"/uploads/test.jpg"},
			CreatedAt:   time.Now(),
			IsVisible:   true,
		}

		mock.ExpectQuery(`INSERT INTO contents`).
			WithArgs(
				content.ID,
				content.AuthorId,
				content.Description,
				pq.Array(content.MediaURLs),
				content.CreatedAt,
				content.IsVisible,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(content.ID))

		contentID, err := repo.Create(ctx, content)
		assert.NoError(t, err)
		assert.Equal(t, content.ID, contentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка при создании контента", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		content := &Content{
			ID:          uuid.New(),
			AuthorId:    uuid.New(),
			Description: "Test",
			MediaURLs:   []string{},
			CreatedAt:   time.Now(),
			IsVisible:   true,
		}

		mock.ExpectQuery(`INSERT INTO contents`).
			WithArgs(
				content.ID,
				content.AuthorId,
				content.Description,
				pq.Array(content.MediaURLs),
				content.CreatedAt,
				content.IsVisible,
			).
			WillReturnError(sql.ErrConnDone)

		contentID, err := repo.Create(ctx, content)
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, contentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContentRepository_GetByID(t *testing.T) {
	t.Run("Успешное получение контента по ID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()
		authorID := uuid.New()
		createdAt := time.Now()
		avatarURL := "/uploads/avatar.jpg"

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		}).AddRow(
			contentID,
			"Test description",
			pq.Array([]string{"/uploads/test.jpg"}),
			createdAt,
			true,
			authorID,
			"testuser",
			avatarURL,
		)

		mock.ExpectQuery(`SELECT`).
			WithArgs(contentID).
			WillReturnRows(rows)

		content, err := repo.GetByID(ctx, contentID)
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.Equal(t, contentID, content.ID)
		assert.Equal(t, "Test description", content.Description)
		assert.Equal(t, authorID, content.Author.ID)
		assert.Equal(t, "testuser", content.Author.Username)
		assert.Equal(t, avatarURL, *content.Author.AvatarURL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Контент не найден", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()

		mock.ExpectQuery(`SELECT`).
			WithArgs(contentID).
			WillReturnError(sql.ErrNoRows)

		content, err := repo.GetByID(ctx, contentID)
		assert.Error(t, err)
		assert.Nil(t, content)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка базы данных при получении контента", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()

		mock.ExpectQuery(`SELECT`).
			WithArgs(contentID).
			WillReturnError(errors.New("database error"))

		content, err := repo.GetByID(ctx, contentID)
		assert.Error(t, err)
		assert.Nil(t, content)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Получение контента с nil avatar_url", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()
		authorID := uuid.New()
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		}).AddRow(
			contentID,
			"Test description",
			pq.Array([]string{"/uploads/test.jpg"}),
			createdAt,
			true,
			authorID,
			"testuser",
			nil,
		)

		mock.ExpectQuery(`SELECT`).
			WithArgs(contentID).
			WillReturnRows(rows)

		content, err := repo.GetByID(ctx, contentID)
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.Equal(t, contentID, content.ID)
		assert.Nil(t, content.Author.AvatarURL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContentRepository_GetByAuthor(t *testing.T) {
	t.Run("Успешное получение контента по автору", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		authorID := uuid.New()
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		}).AddRow(
			uuid.New(),
			"Post 1",
			pq.Array([]string{"/uploads/img1.jpg"}),
			createdAt,
			true,
			authorID,
			"user1",
			"/uploads/avatar1.jpg",
		).AddRow(
			uuid.New(),
			"Post 2",
			pq.Array([]string{}),
			createdAt.Add(-time.Hour),
			false,
			authorID,
			"user1",
			"/uploads/avatar1.jpg",
		)

		mock.ExpectQuery(`SELECT`).
			WithArgs(authorID).
			WillReturnRows(rows)

		contents, err := repo.GetByAuthor(ctx, authorID)
		assert.NoError(t, err)
		assert.Len(t, contents, 2)
		assert.Equal(t, "Post 1", contents[0].Description)
		assert.Equal(t, "Post 2", contents[1].Description)
		assert.True(t, contents[0].IsVisible)
		assert.False(t, contents[1].IsVisible)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("У автора нет контента", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		authorID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		})

		mock.ExpectQuery(`SELECT`).
			WithArgs(authorID).
			WillReturnRows(rows)

		contents, err := repo.GetByAuthor(ctx, authorID)
		assert.NoError(t, err)
		assert.Empty(t, contents)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка базы данных при получении контента автора", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		authorID := uuid.New()

		mock.ExpectQuery(`SELECT`).
			WithArgs(authorID).
			WillReturnError(errors.New("database error"))

		contents, err := repo.GetByAuthor(ctx, authorID)
		assert.Error(t, err)
		assert.Nil(t, contents)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContentRepository_Delete(t *testing.T) {
	t.Run("Успешное удаление контента", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()

		mock.ExpectExec(`UPDATE contents`).
			WithArgs(contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Delete(ctx, contentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Контент не найден для удаления", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()

		mock.ExpectExec(`UPDATE contents`).
			WithArgs(contentID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = repo.Delete(ctx, contentID)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка базы данных при удалении", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()

		mock.ExpectExec(`UPDATE contents`).
			WithArgs(contentID).
			WillReturnError(errors.New("database error"))

		err = repo.Delete(ctx, contentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка при получении количества affected rows", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()

		mock.ExpectExec(`UPDATE contents`).
			WithArgs(contentID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err = repo.Delete(ctx, contentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rows affected")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContentRepository_Update(t *testing.T) {
	t.Run("Успешное обновление всех полей контента", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()
		description := "Updated description"
		isVisible := false
		mediaURLs := []string{"/uploads/new.jpg"}

		mock.ExpectExec(`UPDATE contents SET`).
			WithArgs(
				isVisible,
				description,
				pq.Array(mediaURLs),
				contentID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Update(ctx, &description, &isVisible, &mediaURLs, contentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Успешное обновление только описания", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()
		description := "Updated description"

		mock.ExpectExec(`UPDATE contents SET description`).
			WithArgs(description, contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Update(ctx, &description, nil, nil, contentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Успешное обновление только видимости", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()
		isVisible := false

		mock.ExpectExec(`UPDATE contents SET is_visible`).
			WithArgs(isVisible, contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Update(ctx, nil, &isVisible, nil, contentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Успешное обновление только mediaURLs", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()
		mediaURLs := []string{"/uploads/new1.jpg", "/uploads/new2.jpg"}

		mock.ExpectExec(`UPDATE contents SET media_urls`).
			WithArgs(pq.Array(mediaURLs), contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Update(ctx, nil, nil, &mediaURLs, contentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Обновление без изменений", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()

		err = repo.Update(ctx, nil, nil, nil, contentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка базы данных при обновлении", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		contentID := uuid.New()
		description := "Updated description"

		mock.ExpectExec(`UPDATE contents SET description`).
			WithArgs(description, contentID).
			WillReturnError(errors.New("database error"))

		err = repo.Update(ctx, &description, nil, nil, contentID)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContentRepository_GetByFollowing(t *testing.T) {
	t.Run("Успешное получение контента подписок", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		userID := uuid.New()
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		}).AddRow(
			uuid.New(),
			"Post from followed user",
			pq.Array([]string{"/uploads/img1.jpg"}),
			createdAt,
			true,
			uuid.New(),
			"followeduser",
			"/uploads/avatar.jpg",
		)

		mock.ExpectQuery(`SELECT`).
			WithArgs(userID).
			WillReturnRows(rows)

		contents, err := repo.GetByFollowing(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, contents, 1)
		assert.Equal(t, "Post from followed user", contents[0].Description)
		assert.True(t, contents[0].IsVisible)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Пустая лента подписок", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		userID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		})

		mock.ExpectQuery(`SELECT`).
			WithArgs(userID).
			WillReturnRows(rows)

		contents, err := repo.GetByFollowing(ctx, userID)
		assert.NoError(t, err)
		assert.Empty(t, contents)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка базы данных при получении подписок", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		userID := uuid.New()

		mock.ExpectQuery(`SELECT`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		contents, err := repo.GetByFollowing(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, contents)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Получение нескольких постов от подписок", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		userID := uuid.New()
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		}).AddRow(
			uuid.New(),
			"Post 1",
			pq.Array([]string{"/uploads/img1.jpg"}),
			createdAt,
			true,
			uuid.New(),
			"user1",
			"/uploads/avatar1.jpg",
		).AddRow(
			uuid.New(),
			"Post 2",
			pq.Array([]string{}),
			createdAt.Add(-time.Hour),
			true,
			uuid.New(),
			"user2",
			nil,
		)

		mock.ExpectQuery(`SELECT`).
			WithArgs(userID).
			WillReturnRows(rows)

		contents, err := repo.GetByFollowing(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, contents, 2)
		assert.Equal(t, "Post 1", contents[0].Description)
		assert.Equal(t, "Post 2", contents[1].Description)
		assert.Nil(t, contents[1].Author.AvatarURL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContentRepository_GetAll(t *testing.T) {
	t.Run("Успешное получение всего контента", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		}).AddRow(
			uuid.New(),
			"Public post 1",
			pq.Array([]string{"/uploads/img1.jpg"}),
			createdAt,
			true,
			uuid.New(),
			"user1",
			"/uploads/avatar1.jpg",
		).AddRow(
			uuid.New(),
			"Public post 2",
			pq.Array([]string{}),
			createdAt.Add(-time.Hour),
			true,
			uuid.New(),
			"user2",
			nil,
		)

		mock.ExpectQuery(`SELECT`).
			WillReturnRows(rows)

		contents, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, contents, 2)
		assert.Equal(t, "Public post 1", contents[0].Description)
		assert.Equal(t, "Public post 2", contents[1].Description)
		assert.Nil(t, contents[1].Author.AvatarURL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Пустой список контента", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		})

		mock.ExpectQuery(`SELECT`).
			WillReturnRows(rows)

		contents, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Empty(t, contents)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка базы данных при получении всего контента", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		mock.ExpectQuery(`SELECT`).
			WillReturnError(errors.New("database error"))

		contents, err := repo.GetAll(ctx)
		assert.Error(t, err)
		assert.Nil(t, contents)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Получение контента с разными статусами видимости", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := NewContentRepository(db)
		ctx := context.Background()

		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "description", "media_urls", "created_at", "is_visible",
			"author_id", "username", "avatar_url",
		}).AddRow(
			uuid.New(),
			"Visible post",
			pq.Array([]string{"/uploads/img1.jpg"}),
			createdAt,
			true,
			uuid.New(),
			"user1",
			"/uploads/avatar1.jpg",
		)

		mock.ExpectQuery(`SELECT.*WHERE c.is_visible = true`).
			WillReturnRows(rows)

		contents, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, contents, 1)
		assert.True(t, contents[0].IsVisible)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
