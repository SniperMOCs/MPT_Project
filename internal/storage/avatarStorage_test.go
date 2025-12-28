package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type TestReader struct {
	content   string
	readCount int
	failAfter int
}

func (r *TestReader) Read(p []byte) (n int, err error) {
	r.readCount++
	if r.failAfter > 0 && r.readCount > r.failAfter {
		return 0, errors.New("simulated read error")
	}

	if r.readCount > 1 {
		return 0, io.EOF
	}

	copy(p, []byte(r.content))
	return len(r.content), nil
}

func TestFileAvatarStorage_Save(t *testing.T) {

	tmpDir, err := os.MkdirTemp("", "avatar-storage-test-*")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	userID := uuid.New()
	ctx := context.Background()

	t.Run("Успешное сохранение аватара", func(t *testing.T) {
		storage := NewAvatarStorage(tmpDir, "/uploads")

		imageData := "fake image data"
		reader := &TestReader{content: imageData}

		urlPath, err := storage.Save(ctx, userID, reader)
		if err != nil {
			t.Fatalf("Save() вернула ошибку: %v", err)
		}

		expectedURL := "/uploads/avatars/" + userID.String() + ".png"
		if urlPath != expectedURL {
			t.Errorf("Save() вернула URL = %s, ожидался %s", urlPath, expectedURL)
		}

		fsPath := filepath.Join(tmpDir, "avatars", userID.String()+".png")
		fileInfo, err := os.Stat(fsPath)
		if err != nil {
			t.Fatalf("Файл не создан: %v", err)
		}

		if fileInfo.Size() != int64(len(imageData)) {
			t.Errorf("Размер файла = %d, ожидался %d", fileInfo.Size(), len(imageData))
		}

		content, err := os.ReadFile(fsPath)
		if err != nil {
			t.Fatalf("Не удалось прочитать файл: %v", err)
		}
		if string(content) != imageData {
			t.Errorf("Содержимое файла не совпадает")
		}
	})

	t.Run("Создание директории, если она не существует", func(t *testing.T) {

		newTmpDir, err := os.MkdirTemp("", "avatar-storage-test2-*")
		if err != nil {
			t.Fatalf("Не удалось создать временную директорию: %v", err)
		}
		defer os.RemoveAll(newTmpDir)

		storage := NewAvatarStorage(newTmpDir, "/uploads")

		avatarsDir := filepath.Join(newTmpDir, "avatars")
		if _, err := os.Stat(avatarsDir); !os.IsNotExist(err) {
			t.Fatal("Директория avatars не должна существовать до вызова Save()")
		}

		reader := &TestReader{content: "test"}
		_, err = storage.Save(ctx, userID, reader)
		if err != nil {
			t.Fatalf("Save() вернула ошибку: %v", err)
		}

		if _, err := os.Stat(avatarsDir); err != nil {
			t.Fatalf("Директория avatars не создана: %v", err)
		}
	})

	t.Run("Ошибка при создании директории", func(t *testing.T) {

		badDir := filepath.Join(tmpDir, "bad_dir", "file")
		os.MkdirAll(filepath.Dir(badDir), 0755)
		os.WriteFile(badDir, []byte(""), 0644)

		storage := NewAvatarStorage(badDir, "/uploads")

		reader := &TestReader{content: "test"}
		_, err := storage.Save(ctx, userID, reader)
		if err == nil {
			t.Error("Save() должна вернуть ошибку при невозможности создать директорию")
		}
	})

	t.Run("Ошибка при создании файла", func(t *testing.T) {

		badAvatarsDir := filepath.Join(tmpDir, "bad_avatars")
		os.WriteFile(badAvatarsDir, []byte(""), 0644)

		storage := NewAvatarStorage(badAvatarsDir, "/uploads")

		reader := &TestReader{content: "test"}
		_, err := storage.Save(ctx, userID, reader)
		if err == nil {
			t.Error("Save() должна вернуть ошибку при невозможности создать файл")
		}
	})

	t.Run("Ошибка при копировании данных", func(t *testing.T) {
		storage := NewAvatarStorage(tmpDir, "/uploads")

		reader := &TestReader{
			content:   "test",
			failAfter: 1,
		}

		_, err := storage.Save(ctx, userID, reader)
		if err == nil {
			t.Error("Save() должна вернуть ошибку при ошибке чтения")
		}
	})

	t.Run("Корректные пути для разных userID", func(t *testing.T) {
		storage := NewAvatarStorage(tmpDir, "/static")

		testCases := []struct {
			name   string
			userID uuid.UUID
		}{
			{"UUID 1", uuid.New()},
			{"UUID 2", uuid.New()},
			{"UUID 3", uuid.New()},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				reader := &TestReader{content: "avatar data"}
				urlPath, err := storage.Save(ctx, tc.userID, reader)
				if err != nil {
					t.Fatalf("Save() вернула ошибку: %v", err)
				}

				expectedURL := "/static/avatars/" + tc.userID.String() + ".png"
				if urlPath != expectedURL {
					t.Errorf("Save() вернула URL = %s, ожидался %s", urlPath, expectedURL)
				}

				fsPath := filepath.Join(tmpDir, "avatars", tc.userID.String()+".png")
				if _, err := os.Stat(fsPath); err != nil {
					t.Errorf("Файл для %s не создан: %v", tc.name, err)
				}
			})
		}
	})

	t.Run("Разные base URLs", func(t *testing.T) {
		testCases := []struct {
			name       string
			fsBaseDir  string
			urlBaseDir string
		}{
			{"Простой путь", tmpDir, "/uploads"},
			{"Вложенный путь", tmpDir, "/static/images"},
			{"Корневой путь", tmpDir, "/"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {

				testDir, err := os.MkdirTemp(tmpDir, tc.name+"-*")
				if err != nil {
					t.Fatalf("Не удалось создать директорию: %v", err)
				}

				storage := NewAvatarStorage(testDir, tc.urlBaseDir)
				reader := &TestReader{content: "test"}

				urlPath, err := storage.Save(ctx, userID, reader)
				if err != nil {
					t.Fatalf("Save() вернула ошибку: %v", err)
				}

				expectedURL := tc.urlBaseDir + "/avatars/" + userID.String() + ".png"
				if strings.HasSuffix(tc.urlBaseDir, "/") {
					expectedURL = tc.urlBaseDir + "avatars/" + userID.String() + ".png"
				}

				if urlPath != expectedURL {
					t.Errorf("Save() вернула URL = %s, ожидался %s", urlPath, expectedURL)
				}
			})
		}
	})
}

func TestNewAvatarStorage(t *testing.T) {
	t.Run("Создание storage с разными параметрами", func(t *testing.T) {
		testCases := []struct {
			name       string
			fsBaseDir  string
			urlBaseDir string
		}{
			{"Обычные пути", "/var/www", "/static"},
			{"Вложенные пути", "/var/www/uploads", "/static/images"},
			{"Корневые пути", "/", "/"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				storage := NewAvatarStorage(tc.fsBaseDir, tc.urlBaseDir)

				if storage == nil {
					t.Error("NewAvatarStorage() вернул nil")
				}

				if storage.fsBaseDir != tc.fsBaseDir {
					t.Errorf("fsBaseDir = %s, ожидался %s", storage.fsBaseDir, tc.fsBaseDir)
				}
				if storage.urlBaseDir != tc.urlBaseDir {
					t.Errorf("urlBaseDir = %s, ожидался %s", storage.urlBaseDir, tc.urlBaseDir)
				}
			})
		}
	})
}

func TestFileAvatarStorage_ConcurrentSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "avatar-concurrent-test-*")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewAvatarStorage(tmpDir, "/uploads")
	ctx := context.Background()

	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			userID := uuid.New()
			content := strings.Repeat("x", 100+index)
			reader := &TestReader{content: content}

			_, err := storage.Save(ctx, userID, reader)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		select {
		case <-done:

		case err := <-errors:
			t.Errorf("Ошибка при конкурентном сохранении: %v", err)
		}
	}

	avatarsDir := filepath.Join(tmpDir, "avatars")
	files, err := os.ReadDir(avatarsDir)
	if err != nil {
		t.Fatalf("Не удалось прочитать директорию: %v", err)
	}

	if len(files) != 10 {
		t.Errorf("Создано %d файлов, ожидалось 10", len(files))
	}
}
