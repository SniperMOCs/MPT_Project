package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type FileMediaStorage struct {
	fsBaseDir  string
	urlBaseDir string
}

func NewFileMediaStorage(fsBaseDir string, urlBaseDir string) *FileMediaStorage {
	return &FileMediaStorage{
		fsBaseDir:  fsBaseDir,
		urlBaseDir: urlBaseDir,
	}
}

func (s *FileMediaStorage) Delete(ctx context.Context, urls []string) error {
	for _, url := range urls {
		_ = os.Remove("." + url)
	}
	return nil
}

func (s *FileMediaStorage) Save(ctx context.Context, contentId uuid.UUID, files []*multipart.FileHeader) ([]string, error) {
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
		".mp4": true, ".mov": true, ".webm": true,
	}

	dir := filepath.Join(s.fsBaseDir, "content", contentId.String())
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	var urls []string
	for _, h := range files {
		ext := strings.ToLower(filepath.Ext(h.Filename))
		if !allowed[ext] {
			return nil, fmt.Errorf("unsupported file type: %s", ext)
		}

		file, err := h.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		fileName := uuid.New().String() + ext
		fsPath := filepath.Join(dir, fileName)

		out, err := os.Create(fsPath)
		if err != nil {
			return nil, err
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			return nil, err
		}
		defer out.Close()

		url := path.Join(
			s.urlBaseDir,
			"content",
			contentId.String(),
			fileName,
		)

		urls = append(urls, url)
	}

	return urls, nil
}
