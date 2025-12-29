package comment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestCommentRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepository(db)

	t.Run("Успешное создание комментария", func(t *testing.T) {
		comment := &Comment{
			ID:        uuid.New(),
			AuthorId:  uuid.New(),
			ParentId:  nil,
			ContentId: uuid.New(),
			Text:      "Test comment",
			CreatedAt: time.Now(),
		}

		mock.ExpectExec(`INSERT INTO comments`).
			WithArgs(comment.ID, comment.AuthorId, comment.ParentId, comment.ContentId, comment.Text, comment.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), comment)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Успешное создание ответного комментария", func(t *testing.T) {
		parentID := uuid.New()
		comment := &Comment{
			ID:        uuid.New(),
			AuthorId:  uuid.New(),
			ParentId:  &parentID,
			ContentId: uuid.New(),
			Text:      "Test reply",
			CreatedAt: time.Now(),
		}

		mock.ExpectExec(`INSERT INTO comments`).
			WithArgs(comment.ID, comment.AuthorId, comment.ParentId, comment.ContentId, comment.Text, comment.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), comment)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при создании комментария", func(t *testing.T) {
		comment := &Comment{
			ID:        uuid.New(),
			AuthorId:  uuid.New(),
			ParentId:  nil,
			ContentId: uuid.New(),
			Text:      "Test comment",
			CreatedAt: time.Now(),
		}

		mock.ExpectExec(`INSERT INTO comments`).
			WithArgs(comment.ID, comment.AuthorId, comment.ParentId, comment.ContentId, comment.Text, comment.CreatedAt).
			WillReturnError(errors.New("database error"))

		err := repo.Create(context.Background(), comment)
		if err == nil {
			t.Error("Create() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestCommentRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepository(db)

	t.Run("Успешное удаление комментария", func(t *testing.T) {
		commentID := uuid.New()

		mock.ExpectExec(`DELETE FROM comments WHERE id = \$1`).
			WithArgs(commentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), commentID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при удалении комментария", func(t *testing.T) {
		commentID := uuid.New()

		mock.ExpectExec(`DELETE FROM comments WHERE id = \$1`).
			WithArgs(commentID).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(context.Background(), commentID)
		if err == nil {
			t.Error("Delete() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestCommentRepository_GetByContentId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepository(db)

	t.Run("Успешное получение комментариев к посту", func(t *testing.T) {
		contentID := uuid.New()
		authorID := uuid.New()
		parentID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "parent_id", "content_id", "text", "created_at", "author_id", "username", "avatar_url",
		}).
			AddRow(uuid.New(), &parentID, contentID, "Comment 1", time.Now(), authorID, "user1", "/avatar1.jpg").
			AddRow(uuid.New(), nil, contentID, "Comment 2", time.Now(), authorID, "user2", "/avatar2.jpg")

		mock.ExpectQuery(`SELECT c\.id, c\.parent_id, c\.content_id, c\.text, c\.created_at, c\.author_id, u\.username, u\.avatar_url FROM comments c INNER JOIN users u ON u\.id = c\.author_id WHERE content_id = \$1 ORDER BY created_at ASC`).
			WithArgs(contentID).
			WillReturnRows(rows)

		comments, err := repo.GetByContentId(context.Background(), contentID)
		if err != nil {
			t.Errorf("GetByContentId() error = %v", err)
		}
		if len(comments) != 2 {
			t.Errorf("GetByContentId() вернул %d комментариев, ожидалось 2", len(comments))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Нет комментариев к посту", func(t *testing.T) {
		contentID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "parent_id", "content_id", "text", "created_at", "author_id", "username", "avatar_url",
		})

		mock.ExpectQuery(`SELECT c\.id, c\.parent_id, c\.content_id, c\.text, c\.created_at, c\.author_id, u\.username, u\.avatar_url FROM comments c INNER JOIN users u ON u\.id = c\.author_id WHERE content_id = \$1 ORDER BY created_at ASC`).
			WithArgs(contentID).
			WillReturnRows(rows)

		comments, err := repo.GetByContentId(context.Background(), contentID)
		if err != nil {
			t.Errorf("GetByContentId() error = %v", err)
		}
		if len(comments) != 0 {
			t.Errorf("GetByContentId() вернул %d комментариев, ожидалось 0", len(comments))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при получении комментариев", func(t *testing.T) {
		contentID := uuid.New()

		mock.ExpectQuery(`SELECT c\.id, c\.parent_id, c\.content_id, c\.text, c\.created_at, c\.author_id, u\.username, u\.avatar_url FROM comments c INNER JOIN users u ON u\.id = c\.author_id WHERE content_id = \$1 ORDER BY created_at ASC`).
			WithArgs(contentID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetByContentId(context.Background(), contentID)
		if err == nil {
			t.Error("GetByContentId() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestCommentRepository_GetById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewCommentRepository(db)

	t.Run("Успешное получение комментария по ID", func(t *testing.T) {
		commentID := uuid.New()
		parentID := uuid.New()
		contentID := uuid.New()
		authorID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "parent_id", "content_id", "text", "created_at", "author_id",
		}).AddRow(commentID, &parentID, contentID, "Test comment", time.Now(), authorID)

		mock.ExpectQuery(`SELECT c\.id, c\.parent_id, c\.content_id, c\.text, c\.created_at, c\.author_id FROM comments c WHERE id = \$1`).
			WithArgs(commentID).
			WillReturnRows(rows)

		comment, err := repo.GetById(context.Background(), commentID)
		if err != nil {
			t.Errorf("GetById() error = %v", err)
		}
		if comment.ID != commentID {
			t.Errorf("GetById() вернул комментарий с ID %v, ожидалось %v", comment.ID, commentID)
		}
		if comment.Text != "Test comment" {
			t.Errorf("GetById() вернул текст '%s', ожидалось 'Test comment'", comment.Text)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Комментарий не найден", func(t *testing.T) {
		commentID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "parent_id", "content_id", "text", "created_at", "author_id",
		})

		mock.ExpectQuery(`SELECT c\.id, c\.parent_id, c\.content_id, c\.text, c\.created_at, c\.author_id FROM comments c WHERE id = \$1`).
			WithArgs(commentID).
			WillReturnRows(rows)

		_, err := repo.GetById(context.Background(), commentID)
		if err == nil {
			t.Error("GetById() должен вернуть ошибку для несуществующего комментария")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при получении комментария", func(t *testing.T) {
		commentID := uuid.New()

		mock.ExpectQuery(`SELECT c\.id, c\.parent_id, c\.content_id, c\.text, c\.created_at, c\.author_id FROM comments c WHERE id = \$1`).
			WithArgs(commentID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetById(context.Background(), commentID)
		if err == nil {
			t.Error("GetById() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}
