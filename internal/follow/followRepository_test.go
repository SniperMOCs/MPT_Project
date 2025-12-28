package follow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SniperMOCs/MPT_Project/internal/user"
	"github.com/google/uuid"
)

func TestFollowRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewFollowRepository(db)

	t.Run("Успешное создание подписки", func(t *testing.T) {
		followerID := uuid.New()
		followingID := uuid.New()

		mock.ExpectExec(`INSERT INTO follows`).
			WithArgs(sqlmock.AnyArg(), followerID, followingID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), followerID, followingID)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при создании подписки", func(t *testing.T) {
		followerID := uuid.New()
		followingID := uuid.New()

		mock.ExpectExec(`INSERT INTO follows`).
			WithArgs(sqlmock.AnyArg(), followerID, followingID, sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.Create(context.Background(), followerID, followingID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestFollowRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewFollowRepository(db)

	t.Run("Успешное удаление подписки", func(t *testing.T) {
		followerID := uuid.New()
		followingID := uuid.New()

		mock.ExpectExec(`DELETE FROM follows WHERE follower_id = \$1 AND following_id = \$2`).
			WithArgs(followerID, followingID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), followerID, followingID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при удалении подписки", func(t *testing.T) {
		followerID := uuid.New()
		followingID := uuid.New()

		mock.ExpectExec(`DELETE FROM follows WHERE follower_id = \$1 AND following_id = \$2`).
			WithArgs(followerID, followingID).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(context.Background(), followerID, followingID)
		if err == nil {
			t.Error("Delete() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestFollowRepository_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewFollowRepository(db)

	t.Run("Подписка существует", func(t *testing.T) {
		followerID := uuid.New()
		followingID := uuid.New()

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(followerID, followingID).
			WillReturnRows(rows)

		exists, err := repo.Exists(context.Background(), followerID, followingID)
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

	t.Run("Подписка не существует", func(t *testing.T) {
		followerID := uuid.New()
		followingID := uuid.New()

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(followerID, followingID).
			WillReturnRows(rows)

		exists, err := repo.Exists(context.Background(), followerID, followingID)
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

	t.Run("Ошибка при проверке существования", func(t *testing.T) {
		followerID := uuid.New()
		followingID := uuid.New()

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(followerID, followingID).
			WillReturnError(errors.New("database error"))

		_, err := repo.Exists(context.Background(), followerID, followingID)
		if err == nil {
			t.Error("Exists() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestFollowRepository_GetFollowers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewFollowRepository(db)

	t.Run("Успешное получение подписчиков", func(t *testing.T) {
		userID := uuid.New()
		expectedUsers := []*user.User{
			{
				ID:          uuid.New(),
				Username:    "follower1",
				Password:    "password1",
				FullName:    strPtr("Follower One"),
				AvatarURL:   strPtr("/avatar1.jpg"),
				Description: strPtr("Description 1"),
				CreatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				Username:    "follower2",
				Password:    "password2",
				FullName:    strPtr("Follower Two"),
				AvatarURL:   strPtr("/avatar2.jpg"),
				Description: strPtr("Description 2"),
				CreatedAt:   time.Now(),
			},
		}

		rows := sqlmock.NewRows([]string{
			"id", "username", "password", "full_name", "avatar_url", "description", "created_at",
		})
		for _, u := range expectedUsers {
			rows.AddRow(
				u.ID, u.Username, u.Password, u.FullName, u.AvatarURL, u.Description, u.CreatedAt,
			)
		}

		mock.ExpectQuery(`SELECT u\.id, u\.username, u\.password, u\.full_name, u\.avatar_url, u\.description, u\.created_at FROM users u INNER JOIN follows f ON u\.id = f\.follower_id WHERE f\.following_id = \$1 ORDER BY f\.created_at DESC`).
			WithArgs(userID).
			WillReturnRows(rows)

		users, err := repo.GetFollowers(context.Background(), userID)
		if err != nil {
			t.Errorf("GetFollowers() error = %v", err)
		}
		if len(users) != len(expectedUsers) {
			t.Errorf("GetFollowers() вернул %d пользователей, ожидалось %d", len(users), len(expectedUsers))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Нет подписчиков", func(t *testing.T) {
		userID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "username", "password", "full_name", "avatar_url", "description", "created_at",
		})

		mock.ExpectQuery(`SELECT u\.id, u\.username, u\.password, u\.full_name, u\.avatar_url, u\.description, u\.created_at FROM users u INNER JOIN follows f ON u\.id = f\.follower_id WHERE f\.following_id = \$1 ORDER BY f\.created_at DESC`).
			WithArgs(userID).
			WillReturnRows(rows)

		users, err := repo.GetFollowers(context.Background(), userID)
		if err != nil {
			t.Errorf("GetFollowers() error = %v", err)
		}
		if len(users) != 0 {
			t.Errorf("GetFollowers() вернул %d пользователей, ожидалось 0", len(users))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при получении подписчиков", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectQuery(`SELECT u\.id, u\.username, u\.password, u\.full_name, u\.avatar_url, u\.description, u\.created_at FROM users u INNER JOIN follows f ON u\.id = f\.follower_id WHERE f\.following_id = \$1 ORDER BY f\.created_at DESC`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetFollowers(context.Background(), userID)
		if err == nil {
			t.Error("GetFollowers() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestFollowRepository_GetFollowing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewFollowRepository(db)

	t.Run("Успешное получение подписок", func(t *testing.T) {
		userID := uuid.New()
		expectedUsers := []*user.User{
			{
				ID:          uuid.New(),
				Username:    "following1",
				Password:    "password1",
				FullName:    strPtr("Following One"),
				AvatarURL:   strPtr("/avatar1.jpg"),
				Description: strPtr("Description 1"),
				CreatedAt:   time.Now(),
			},
		}

		rows := sqlmock.NewRows([]string{
			"id", "username", "password", "full_name", "avatar_url", "description", "created_at",
		})
		for _, u := range expectedUsers {
			rows.AddRow(
				u.ID, u.Username, u.Password, u.FullName, u.AvatarURL, u.Description, u.CreatedAt,
			)
		}

		mock.ExpectQuery(`SELECT u\.id, u\.username, u\.password, u\.full_name, u\.avatar_url, u\.description, u\.created_at FROM users u INNER JOIN follows f ON u\.id = f\.following_id WHERE f\.follower_id = \$1 ORDER BY f\.created_at DESC`).
			WithArgs(userID).
			WillReturnRows(rows)

		users, err := repo.GetFollowing(context.Background(), userID)
		if err != nil {
			t.Errorf("GetFollowing() error = %v", err)
		}
		if len(users) != len(expectedUsers) {
			t.Errorf("GetFollowing() вернул %d пользователей, ожидалось %d", len(users), len(expectedUsers))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при получении подписок", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectQuery(`SELECT u\.id, u\.username, u\.password, u\.full_name, u\.avatar_url, u\.description, u\.created_at FROM users u INNER JOIN follows f ON u\.id = f\.following_id WHERE f\.follower_id = \$1 ORDER BY f\.created_at DESC`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetFollowing(context.Background(), userID)
		if err == nil {
			t.Error("GetFollowing() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestFollowRepository_GetFollowersCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewFollowRepository(db)

	t.Run("Успешное получение количества подписчиков", func(t *testing.T) {
		userID := uuid.New()
		expectedCount := 42

		rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedCount)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		count, err := repo.GetFollowersCount(context.Background(), userID)
		if err != nil {
			t.Errorf("GetFollowersCount() error = %v", err)
		}
		if count != expectedCount {
			t.Errorf("GetFollowersCount() = %v, ожидалось %v", count, expectedCount)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при получении количества подписчиков", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id = \$1`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetFollowersCount(context.Background(), userID)
		if err == nil {
			t.Error("GetFollowersCount() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestFollowRepository_GetFollowingCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewFollowRepository(db)

	t.Run("Успешное получение количества подписок", func(t *testing.T) {
		userID := uuid.New()
		expectedCount := 15

		rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedCount)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE follower_id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		count, err := repo.GetFollowingCount(context.Background(), userID)
		if err != nil {
			t.Errorf("GetFollowingCount() error = %v", err)
		}
		if count != expectedCount {
			t.Errorf("GetFollowingCount() = %v, ожидалось %v", count, expectedCount)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при получении количества подписок", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE follower_id = \$1`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetFollowingCount(context.Background(), userID)
		if err == nil {
			t.Error("GetFollowingCount() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func strPtr(s string) *string {
	return &s
}
