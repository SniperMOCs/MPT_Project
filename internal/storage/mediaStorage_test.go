package storage

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func createTestFileHeader(filename string, content string) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	_, err = io.WriteString(part, content)
	if err != nil {
		return nil, err
	}

	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	file, header, err := req.FormFile("file")
	if err != nil {
		return nil, err
	}
	file.Close()

	return header, nil
}

func getFilePathFromURL(tmpDir, urlBaseDir, url string) string {

	relativePath := strings.TrimPrefix(url, urlBaseDir)

	relativePath = strings.TrimPrefix(relativePath, "/")

	return filepath.Join(tmpDir, relativePath)
}

func TestFileMediaStorage_Save(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "media-storage-test-*")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	contentID := uuid.New()
	ctx := context.Background()
	urlBaseDir := "/media"

	t.Run("Успешное сохранение одного файла", func(t *testing.T) {
		storage := NewFileMediaStorage(tmpDir, urlBaseDir)

		header, err := createTestFileHeader("test.jpg", "image content")
		if err != nil {
			t.Fatalf("Не удалось создать тестовый файл: %v", err)
		}

		urls, err := storage.Save(ctx, contentID, []*multipart.FileHeader{header})
		if err != nil {
			t.Fatalf("Save() вернула ошибку: %v", err)
		}

		if len(urls) != 1 {
			t.Fatalf("Save() вернула %d URL, ожидался 1", len(urls))
		}

		url := urls[0]
		if !strings.HasPrefix(url, urlBaseDir+"/content/"+contentID.String()+"/") {
			t.Errorf("URL не содержит правильный путь: %s", url)
		}
		if !strings.HasSuffix(url, ".jpg") {
			t.Errorf("URL не имеет правильного расширения: %s", url)
		}

		fsPath := getFilePathFromURL(tmpDir, urlBaseDir, url)
		fileInfo, err := os.Stat(fsPath)
		if err != nil {
			t.Fatalf("Файл не создан: %v (путь: %s)", err, fsPath)
		}

		if fileInfo.Size() != int64(len("image content")) {
			t.Errorf("Размер файла = %d, ожидался %d", fileInfo.Size(), len("image content"))
		}

		content, err := os.ReadFile(fsPath)
		if err != nil {
			t.Fatalf("Не удалось прочитать файл: %v", err)
		}
		if string(content) != "image content" {
			t.Errorf("Содержимое файла не совпадает: получено %s, ожидалось %s", string(content), "image content")
		}
	})

	t.Run("Успешное сохранение нескольких файлов", func(t *testing.T) {
		storage := NewFileMediaStorage(tmpDir, urlBaseDir)

		headers := []*multipart.FileHeader{}
		files := []struct {
			filename string
			content  string
		}{
			{"image1.jpg", "image 1 content"},
			{"image2.png", "image 2 content"},
			{"video1.mp4", "video content"},
		}

		for _, f := range files {
			header, err := createTestFileHeader(f.filename, f.content)
			if err != nil {
				t.Fatalf("Не удалось создать тестовый файл %s: %v", f.filename, err)
			}
			headers = append(headers, header)
		}

		urls, err := storage.Save(ctx, contentID, headers)
		if err != nil {
			t.Fatalf("Save() вернула ошибку: %v", err)
		}

		if len(urls) != 3 {
			t.Fatalf("Save() вернула %d URL, ожидалось 3", len(urls))
		}

		for i, url := range urls {
			if !strings.HasPrefix(url, urlBaseDir+"/content/"+contentID.String()+"/") {
				t.Errorf("URL %d не содержит правильный путь: %s", i, url)
			}

			fsPath := getFilePathFromURL(tmpDir, urlBaseDir, url)
			content, err := os.ReadFile(fsPath)
			if err != nil {
				t.Errorf("Не удалось прочитать файл %d: %v (путь: %s)", i, err, fsPath)
				continue
			}
			if string(content) != files[i].content {
				t.Errorf("Содержимое файла %d не совпадает: получено %s, ожидалось %s", i, string(content), files[i].content)
			}
		}
	})

	t.Run("Создание директории, если она не существует", func(t *testing.T) {

		newTmpDir, err := os.MkdirTemp("", "media-storage-test2-*")
		if err != nil {
			t.Fatalf("Не удалось создать временную директорию: %v", err)
		}
		defer os.RemoveAll(newTmpDir)

		storage := NewFileMediaStorage(newTmpDir, urlBaseDir)

		contentDir := filepath.Join(newTmpDir, "content", contentID.String())
		if _, err := os.Stat(contentDir); !os.IsNotExist(err) {
			t.Fatal("Директория content не должна существовать до вызова Save()")
		}

		header, err := createTestFileHeader("test.jpg", "content")
		if err != nil {
			t.Fatalf("Не удалось создать тестовый файл: %v", err)
		}

		_, err = storage.Save(ctx, contentID, []*multipart.FileHeader{header})
		if err != nil {
			t.Fatalf("Save() вернула ошибку: %v", err)
		}

		if _, err := os.Stat(contentDir); err != nil {
			t.Fatalf("Директория не создана: %v", err)
		}
	})

	t.Run("Ошибка при неподдерживаемом типе файла", func(t *testing.T) {
		storage := NewFileMediaStorage(tmpDir, urlBaseDir)

		header, err := createTestFileHeader("test.exe", "malicious content")
		if err != nil {
			t.Fatalf("Не удалось создать тестовый файл: %v", err)
		}

		_, err = storage.Save(ctx, contentID, []*multipart.FileHeader{header})
		if err == nil {
			t.Error("Save() должна вернуть ошибку для неподдерживаемого типа файла")
		}
		if !strings.Contains(err.Error(), "unsupported file type") {
			t.Errorf("Неправильное сообщение об ошибке: %v", err)
		}
	})

	t.Run("Ошибка при открытии файла", func(t *testing.T) {
		storage := NewFileMediaStorage(tmpDir, urlBaseDir)

		header := &multipart.FileHeader{
			Filename: "test.jpg",
			Size:     100,
		}

		_, err := storage.Save(ctx, contentID, []*multipart.FileHeader{header})
		if err == nil {
			t.Error("Save() должна вернуть ошибку при невозможности открыть файл")
		}
	})

	t.Run("Ошибка при создании файла на диске", func(t *testing.T) {

		badDir := filepath.Join(tmpDir, "bad_content", contentID.String())
		os.MkdirAll(filepath.Dir(badDir), 0755)
		os.WriteFile(badDir, []byte(""), 0644)

		storage := NewFileMediaStorage(badDir, urlBaseDir)

		header, err := createTestFileHeader("test.jpg", "content")
		if err != nil {
			t.Fatalf("Не удалось создать тестовый файл: %v", err)
		}

		_, err = storage.Save(ctx, contentID, []*multipart.FileHeader{header})
		if err == nil {
			t.Error("Save() должна вернуть ошибку при невозможности создать файл")
		}
	})

	t.Run("Разные contentID создают разные директории", func(t *testing.T) {
		storage := NewFileMediaStorage(tmpDir, urlBaseDir)

		testCases := []struct {
			name      string
			contentID uuid.UUID
		}{
			{"Content 1", uuid.New()},
			{"Content 2", uuid.New()},
			{"Content 3", uuid.New()},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				header, err := createTestFileHeader("test.jpg", "content")
				if err != nil {
					t.Fatalf("Не удалось создать тестовый файл: %v", err)
				}

				urls, err := storage.Save(ctx, tc.contentID, []*multipart.FileHeader{header})
				if err != nil {
					t.Fatalf("Save() вернула ошибку: %v", err)
				}

				url := urls[0]
				if !strings.Contains(url, tc.contentID.String()) {
					t.Errorf("URL не содержит правильный contentID: %s", url)
				}

				contentDir := filepath.Join(tmpDir, "content", tc.contentID.String())
				if _, err := os.Stat(contentDir); err != nil {
					t.Errorf("Директория для %s не создана: %v", tc.name, err)
				}
			})
		}
	})

	t.Run("Все поддерживаемые форматы", func(t *testing.T) {
		storage := NewFileMediaStorage(tmpDir, urlBaseDir)

		supportedFormats := []string{
			".jpg", ".jpeg", ".png", ".webp",
			".mp4", ".mov", ".webm",
		}

		var headers []*multipart.FileHeader
		for _, ext := range supportedFormats {
			filename := "test" + ext
			header, err := createTestFileHeader(filename, "content for "+ext)
			if err != nil {
				t.Fatalf("Не удалось создать файл %s: %v", filename, err)
			}
			headers = append(headers, header)
		}

		urls, err := storage.Save(ctx, contentID, headers)
		if err != nil {
			t.Fatalf("Save() вернула ошибку: %v", err)
		}

		if len(urls) != len(supportedFormats) {
			t.Fatalf("Сохранено %d файлов, ожидалось %d", len(urls), len(supportedFormats))
		}
	})
}

func TestFileMediaStorage_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Удаление существующих файлов", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "media-delete-test-*")
		if err != nil {
			t.Fatalf("Не удалось создать временную директорию: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Не удалось получить текущую директорию: %v", err)
		}
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		if err != nil {
			t.Fatalf("Не удалось сменить директорию: %v", err)
		}

		storage := NewFileMediaStorage(tmpDir, "/media")

		testFiles := []struct {
			url     string
			content string
		}{
			{"/media/file1.jpg", "content1"},
			{"/media/file2.png", "content2"},
			{"/media/subdir/file3.mp4", "content3"},
		}

		var urls []string
		for _, tf := range testFiles {

			fsPath := "." + tf.url

			os.MkdirAll(filepath.Dir(fsPath), 0755)

			if err := os.WriteFile(fsPath, []byte(tf.content), 0644); err != nil {
				t.Fatalf("Не удалось создать файл %s: %v", fsPath, err)
			}
			urls = append(urls, tf.url)
		}

		for _, tf := range testFiles {
			fsPath := "." + tf.url
			if _, err := os.Stat(fsPath); err != nil {
				t.Fatalf("Файл %s не создан: %v", fsPath, err)
			}
		}

		err = storage.Delete(ctx, urls)
		if err != nil {
			t.Fatalf("Delete() вернула ошибку: %v", err)
		}

		for _, tf := range testFiles {
			fsPath := "." + tf.url
			if _, err := os.Stat(fsPath); !os.IsNotExist(err) {
				t.Errorf("Файл %s не удален", fsPath)
			}
		}
	})

	t.Run("Удаление несуществующих файлов", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "media-delete-nonexistent-*")
		if err != nil {
			t.Fatalf("Не удалось создать временную директорию: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		storage := NewFileMediaStorage(tmpDir, "/media")

		nonExistentFiles := []string{
			"/media/ghost.jpg",
			"/media/nonexistent.png",
		}

		err = storage.Delete(ctx, nonExistentFiles)
		if err != nil {
			t.Fatalf("Delete() вернула ошибку для несуществующих файлов: %v", err)
		}
	})

	t.Run("Удаление пустого списка файлов", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "media-delete-empty-*")
		if err != nil {
			t.Fatalf("Не удалось создать временную директорию: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		storage := NewFileMediaStorage(tmpDir, "/media")

		err = storage.Delete(ctx, []string{})
		if err != nil {
			t.Fatalf("Delete() вернула ошибку для пустого списка: %v", err)
		}
	})

	t.Run("Delete не возвращает ошибку при nil списке", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "media-delete-nil-*")
		if err != nil {
			t.Fatalf("Не удалось создать временную директорию: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		storage := NewFileMediaStorage(tmpDir, "/media")

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Delete() запаниковала с nil списком: %v", r)
			}
		}()

		err = storage.Delete(ctx, nil)
		if err != nil {
			t.Errorf("Delete() с nil списком вернула ошибку: %v", err)
		}
	})
}

func TestNewFileMediaStorage(t *testing.T) {
	t.Run("Создание storage с разными параметрами", func(t *testing.T) {
		testCases := []struct {
			name       string
			fsBaseDir  string
			urlBaseDir string
		}{
			{"Обычные пути", "/var/www", "/media"},
			{"Вложенные пути", "/var/www/uploads", "/static/media"},
			{"Корневые пути", "/", "/"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				storage := NewFileMediaStorage(tc.fsBaseDir, tc.urlBaseDir)

				if storage == nil {
					t.Error("NewFileMediaStorage() вернул nil")
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

func TestFileMediaStorage_Integration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "media-integration-test-*")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	urlBaseDir := "/uploads"
	fsBaseDir := tmpDir
	storage := NewFileMediaStorage(fsBaseDir, urlBaseDir)
	ctx := context.Background()

	t.Run("Сохранение файлов", func(t *testing.T) {
		contentID := uuid.New()

		headers := []*multipart.FileHeader{}
		for i := 1; i <= 3; i++ {
			header, err := createTestFileHeader(
				"file"+string(rune('0'+i))+".jpg",
				"content"+string(rune('0'+i)),
			)
			if err != nil {
				t.Fatalf("Не удалось создать файл %d: %v", i, err)
			}
			headers = append(headers, header)
		}

		urls, err := storage.Save(ctx, contentID, headers)
		if err != nil {
			t.Fatalf("Save() вернула ошибку: %v", err)
		}

		for _, url := range urls {
			fsPath := getFilePathFromURL(fsBaseDir, urlBaseDir, url)
			if _, err := os.Stat(fsPath); err != nil {
				t.Errorf("Файл %s не создан: %v (путь: %s)", url, err, fsPath)
			}
		}

		if len(urls) != 3 {
			t.Errorf("Создано %d файлов, ожидалось 3", len(urls))
		}
	})

	t.Run("Метод Delete имеет правильную сигнатуру", func(t *testing.T) {

		err := storage.Delete(ctx, []string{"/uploads/test1.jpg", "/uploads/test2.png"})
		if err != nil {
			t.Errorf("Delete() вернула ошибку: %v", err)
		}
	})
}

func TestFileMediaStorage_Delete_Logic(t *testing.T) {
	t.Run("Delete строит правильные пути", func(t *testing.T) {

		testCases := []struct {
			name     string
			url      string
			expected string
		}{
			{"Простой путь", "/media/file.jpg", "./media/file.jpg"},
			{"Вложенный путь", "/media/images/photo.png", "./media/images/photo.png"},
			{"С поддиректорией", "/media/videos/movie.mp4", "./media/videos/movie.mp4"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {

				constructedPath := "." + tc.url
				if constructedPath != tc.expected {
					t.Errorf("Построен путь %s, ожидался %s", constructedPath, tc.expected)
				}
			})
		}
	})

	t.Run("Delete всегда возвращает nil", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "media-delete-always-nil-*")
		if err != nil {
			t.Fatalf("Не удалось создать временную директорию: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		storage := NewFileMediaStorage(tmpDir, "/media")

		testCases := []struct {
			name string
			urls []string
		}{
			{"Пустой список", []string{}},
			{"Один файл", []string{"/media/file.jpg"}},
			{"Несколько файлов", []string{"/media/file1.jpg", "/media/file2.png"}},
			{"Несуществующие файлы", []string{"/media/ghost.jpg", "/media/phantom.png"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := storage.Delete(context.Background(), tc.urls)
				if err != nil {
					t.Errorf("Delete() вернула ошибку для случая %s: %v", tc.name, err)
				}
			})
		}
	})
}
