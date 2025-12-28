package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("Успешное создание пользователя", func(t *testing.T) {
		userID := uuid.New()
		createdAt := time.Now()
		user := &User{
			ID:        userID,
			Username:  "testuser",
			Password:  "hashedpassword",
			CreatedAt: createdAt,
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(userID, "testuser", "hashedpassword", createdAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), user)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при создании пользователя", func(t *testing.T) {
		userID := uuid.New()
		createdAt := time.Now()
		user := &User{
			ID:        userID,
			Username:  "testuser",
			Password:  "hashedpassword",
			CreatedAt: createdAt,
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(userID, "testuser", "hashedpassword", createdAt).
			WillReturnError(errors.New("database error"))

		err := repo.Create(context.Background(), user)
		if err == nil {
			t.Error("Create() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("Успешное получение пользователя по ID", func(t *testing.T) {
		userID := uuid.New()
		createdAt := time.Now()
		fullName := "Test User"
		description := "Test Description"
		avatarURL := "/uploads/avatar.jpg"

		rows := sqlmock.NewRows([]string{"id", "username", "password", "full_name", "avatar_url", "description", "created_at"}).
			AddRow(userID, "testuser", "hashedpassword", fullName, avatarURL, description, createdAt)

		mock.ExpectQuery("SELECT id, username, password, full_name, avatar_url, description, created_at").
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetByID(context.Background(), userID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
		}

		if user.ID != userID {
			t.Errorf("GetByID() вернул ID = %v, ожидался %v", user.ID, userID)
		}
		if user.Username != "testuser" {
			t.Errorf("GetByID() вернул Username = %v, ожидался %v", user.Username, "testuser")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при получении пользователя по ID", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectQuery("SELECT id, username, password, full_name, avatar_url, description, created_at").
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetByID(context.Background(), userID)
		if err == nil {
			t.Error("GetByID() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Пользователь не найден", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectQuery("SELECT id, username, password, full_name, avatar_url, description, created_at").
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetByID(context.Background(), userID)
		if err == nil {
			t.Error("GetByID() должен вернуть ошибку при отсутствии пользователя")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("Успешное получение пользователя по username", func(t *testing.T) {
		userID := uuid.New()
		createdAt := time.Now()
		fullName := "Test User"
		description := "Test Description"
		avatarURL := "/uploads/avatar.jpg"

		rows := sqlmock.NewRows([]string{"id", "username", "password", "full_name", "avatar_url", "description", "created_at"}).
			AddRow(userID, "testuser", "hashedpassword", fullName, avatarURL, description, createdAt)

		mock.ExpectQuery("SELECT id, username, password, full_name, avatar_url, description, created_at").
			WithArgs("testuser").
			WillReturnRows(rows)

		user, err := repo.GetByUsername(context.Background(), "testuser")
		if err != nil {
			t.Errorf("GetByUsername() error = %v", err)
		}

		if user.Username != "testuser" {
			t.Errorf("GetByUsername() вернул Username = %v, ожидался %v", user.Username, "testuser")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при получении пользователя по username", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, password, full_name, avatar_url, description, created_at").
			WithArgs("testuser").
			WillReturnError(errors.New("database error"))

		_, err := repo.GetByUsername(context.Background(), "testuser")
		if err == nil {
			t.Error("GetByUsername() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestUserRepository_ExistsByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("Пользователь существует", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("testuser").
			WillReturnRows(rows)

		exists, err := repo.ExistsByUsername(context.Background(), "testuser")
		if err != nil {
			t.Errorf("ExistsByUsername() error = %v", err)
		}
		if !exists {
			t.Error("ExistsByUsername() должен вернуть true")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Пользователь не существует", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("testuser").
			WillReturnRows(rows)

		exists, err := repo.ExistsByUsername(context.Background(), "testuser")
		if err != nil {
			t.Errorf("ExistsByUsername() error = %v", err)
		}
		if exists {
			t.Error("ExistsByUsername() должен вернуть false")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при проверке существования пользователя", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("testuser").
			WillReturnError(errors.New("database error"))

		_, err := repo.ExistsByUsername(context.Background(), "testuser")
		if err == nil {
			t.Error("ExistsByUsername() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("Успешное обновление всех полей", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("UPDATE users SET username = \\$1, full_name = \\$2, description = \\$3 WHERE id = \\$4").
			WithArgs("newuser", "New Full Name", "New Description", userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), userID, "newuser", "New Full Name", "New Description")
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Обновление только username", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("UPDATE users SET username = \\$1 WHERE id = \\$2").
			WithArgs("newuser", userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), userID, "newuser", "", "")
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Обновление только fullname", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("UPDATE users SET full_name = \\$1 WHERE id = \\$2").
			WithArgs("New Full Name", userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), userID, "", "New Full Name", "")
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Обновление только description", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("UPDATE users SET description = \\$1 WHERE id = \\$2").
			WithArgs("New Description", userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), userID, "", "", "New Description")
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Обновление без изменений", func(t *testing.T) {
		userID := uuid.New()

		// Не должно быть запроса к БД
		err := repo.Update(context.Background(), userID, "", "", "")
		if err != nil {
			t.Errorf("Update() с пустыми полями не должен возвращать ошибку: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при обновлении", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("UPDATE users SET username = \\$1 WHERE id = \\$2").
			WithArgs("newuser", userID).
			WillReturnError(errors.New("database error"))

		err := repo.Update(context.Background(), userID, "newuser", "", "")
		if err == nil {
			t.Error("Update() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("Успешное обновление пароля", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("UPDATE users SET password = \\$1 WHERE id = \\$2").
			WithArgs("newhashedpassword", userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdatePassword(context.Background(), userID, "newhashedpassword")
		if err != nil {
			t.Errorf("UpdatePassword() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при обновлении пароля", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("UPDATE users SET password = \\$1 WHERE id = \\$2").
			WithArgs("newhashedpassword", userID).
			WillReturnError(errors.New("database error"))

		err := repo.UpdatePassword(context.Background(), userID, "newhashedpassword")
		if err == nil {
			t.Error("UpdatePassword() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestUserRepository_UpdateAvatar(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("Успешное обновление аватара", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("UPDATE users SET avatar_url = \\$1 WHERE id = \\$2").
			WithArgs("/uploads/avatar.jpg", userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateAvatar(context.Background(), userID, "/uploads/avatar.jpg")
		if err != nil {
			t.Errorf("UpdateAvatar() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Ошибка при обновлении аватара", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("UPDATE users SET avatar_url = \\$1 WHERE id = \\$2").
			WithArgs("/uploads/avatar.jpg", userID).
			WillReturnError(errors.New("database error"))

		err := repo.UpdateAvatar(context.Background(), userID, "/uploads/avatar.jpg")
		if err == nil {
			t.Error("UpdateAvatar() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}

func TestUserRepository_SearchByFullName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ошибка создания мока БД: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	t.Run("Успешный поиск пользователей", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()
		createdAt := time.Now()
		fullName1 := "John Doe"
		fullName2 := "Jane Smith"

		rows := sqlmock.NewRows([]string{"id", "username", "full_name", "avatar_url", "description", "created_at"}).
			AddRow(userID1, "johndoe", fullName1, "/uploads/avatar1.jpg", "Description 1", createdAt).
			AddRow(userID2, "janesmith", fullName2, "/uploads/avatar2.jpg", "Description 2", createdAt)

		mock.ExpectQuery("SELECT id, username, full_name, avatar_url, description, created_at").
			WillReturnRows(rows)

		users, err := repo.SearchByFullName(context.Background(), "john jane")
		if err != nil {
			t.Errorf("SearchByFullName() error = %v", err)
		}

		if len(users) != 2 {
			t.Errorf("SearchByFullName() вернул %d пользователей, ожидалось 2", len(users))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Поиск с пустым запросом", func(t *testing.T) {
		users, err := repo.SearchByFullName(context.Background(), "")
		if err != nil {
			t.Errorf("SearchByFullName() с пустым запросом не должен возвращать ошибку: %v", err)
		}

		if len(users) != 0 {
			t.Errorf("SearchByFullName() с пустым запросом должен вернуть пустой список")
		}
	})

	t.Run("Ошибка при поиске пользователей", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, full_name, avatar_url, description, created_at").
			WillReturnError(errors.New("database error"))

		_, err := repo.SearchByFullName(context.Background(), "search")
		if err == nil {
			t.Error("SearchByFullName() должен вернуть ошибку")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})

	t.Run("Поиск с одним словом", func(t *testing.T) {
		userID := uuid.New()
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{"id", "username", "full_name", "avatar_url", "description", "created_at"}).
			AddRow(userID, "testuser", "Test User", "/uploads/avatar.jpg", "Description", createdAt)

		mock.ExpectQuery("SELECT id, username, full_name, avatar_url, description, created_at").
			WillReturnRows(rows)

		users, err := repo.SearchByFullName(context.Background(), "test")
		if err != nil {
			t.Errorf("SearchByFullName() error = %v", err)
		}

		if len(users) != 1 {
			t.Errorf("SearchByFullName() вернул %d пользователей, ожидалось 1", len(users))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("невыполненные ожидания: %v", err)
		}
	})
}
