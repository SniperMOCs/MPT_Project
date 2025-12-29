package like

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestLikeRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewLikeRepository(db)

	t.Run("Успешное создание лайка", func(t *testing.T) {
		userID := uuid.New()
		contentID := uuid.New()

		mock.ExpectExec("INSERT INTO likes").
			WithArgs(sqlmock.AnyArg(), userID, contentID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при создании лайка", func(t *testing.T) {
		userID := uuid.New()
		contentID := uuid.New()

		mock.ExpectExec("INSERT INTO likes").
			WithArgs(sqlmock.AnyArg(), userID, contentID, sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.Create(context.Background(), userID, contentID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestLikeRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewLikeRepository(db)

	t.Run("Успешное удаление лайка", func(t *testing.T) {
		userID := uuid.New()
		contentID := uuid.New()

		mock.ExpectExec("DELETE FROM likes WHERE user_id = \\$1 AND content_id = \\$2").
			WithArgs(userID, contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при удалении лайка", func(t *testing.T) {
		userID := uuid.New()
		contentID := uuid.New()

		mock.ExpectExec("DELETE FROM likes WHERE user_id = \\$1 AND content_id = \\$2").
			WithArgs(userID, contentID).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(context.Background(), userID, contentID)
		if err == nil {
			t.Error("Delete() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestLikeRepository_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewLikeRepository(db)

	t.Run("Лайк существует", func(t *testing.T) {
		userID := uuid.New()
		contentID := uuid.New()

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(userID, contentID).
			WillReturnRows(rows)

		exists, err := repo.Exists(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("Exists() должен вернуть true")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Лайк не существует", func(t *testing.T) {
		userID := uuid.New()
		contentID := uuid.New()

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(userID, contentID).
			WillReturnRows(rows)

		exists, err := repo.Exists(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		}
		if exists {
			t.Error("Exists() должен вернуть false")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при проверке существования лайка", func(t *testing.T) {
		userID := uuid.New()
		contentID := uuid.New()

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(userID, contentID).
			WillReturnError(errors.New("database error"))

		_, err := repo.Exists(context.Background(), userID, contentID)
		if err == nil {
			t.Error("Exists() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestLikeRepository_GetLikesCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewLikeRepository(db)

	t.Run("Успешное получение количества лайков", func(t *testing.T) {
		contentID := uuid.New()
		expectedCount := 5

		rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedCount)

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM likes WHERE content_id = \\$1").
			WithArgs(contentID).
			WillReturnRows(rows)

		count, err := repo.GetLikesCount(context.Background(), contentID)
		if err != nil {
			t.Errorf("GetLikesCount() error = %v", err)
		}
		if count != expectedCount {
			t.Errorf("GetLikesCount() = %v, ожидалось %v", count, expectedCount)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Количество лайков равно 0", func(t *testing.T) {
		contentID := uuid.New()

		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM likes WHERE content_id = \\$1").
			WithArgs(contentID).
			WillReturnRows(rows)

		count, err := repo.GetLikesCount(context.Background(), contentID)
		if err != nil {
			t.Errorf("GetLikesCount() error = %v", err)
		}
		if count != 0 {
			t.Errorf("GetLikesCount() = %v, ожидалось 0", count)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при получении количества лайков", func(t *testing.T) {
		contentID := uuid.New()

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM likes WHERE content_id = \\$1").
			WithArgs(contentID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetLikesCount(context.Background(), contentID)
		if err == nil {
			t.Error("GetLikesCount() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Некорректный результат сканирования", func(t *testing.T) {
		contentID := uuid.New()

		rows := sqlmock.NewRows([]string{"count"}).AddRow("not_a_number")

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM likes WHERE content_id = \\$1").
			WithArgs(contentID).
			WillReturnRows(rows)

		_, err := repo.GetLikesCount(context.Background(), contentID)
		if err == nil {
			t.Error("GetLikesCount() должен вернуть ошибку при некорректном результате")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestLikeRepository_EdgeCases(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewLikeRepository(db)

	t.Run("Удаление несуществующего лайка", func(t *testing.T) {
		userID := uuid.New()
		contentID := uuid.New()

		mock.ExpectExec("DELETE FROM likes WHERE user_id = \\$1 AND content_id = \\$2").
			WithArgs(userID, contentID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Delete() несуществующего лайка error = %v, ожидалось nil", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}
