package storage

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
)

type FileAvatarStorage struct {
	fsBaseDir  string
	urlBaseDir string
}

func NewAvatarStorage(fsBaseDir string, urlBaseDir string) *FileAvatarStorage {
	return &FileAvatarStorage{
		fsBaseDir:  fsBaseDir,
		urlBaseDir: urlBaseDir,
	}
}

func (s *FileAvatarStorage) Save(ctx context.Context, userID uuid.UUID, r io.Reader) (string, error) {
	fileName := userID.String() + ".png"

	fsDir := filepath.Join(s.fsBaseDir, "avatars")
	if err := os.MkdirAll(fsDir, os.ModePerm); err != nil {
		return "", err
	}

	fsPath := filepath.Join(fsDir, fileName)

	out, err := os.Create(fsPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		return "", err
	}

	urlPath := path.Join(s.urlBaseDir, "avatars", fileName)

	return urlPath, nil
}
