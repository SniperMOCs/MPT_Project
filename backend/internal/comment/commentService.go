package comment

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type commentService struct {
	repo CommentRepository
}

func NewCommentService(repo CommentRepository) CommentService {
	return &commentService{repo: repo}
}

func (s *commentService) Create(ctx context.Context, authorId uuid.UUID, parentId *uuid.UUID, contentId uuid.UUID, text string) error {
	comment := &Comment{
		ID:        uuid.New(),
		AuthorId:  authorId,
		ParentId:  parentId,
		ContentId: contentId,
		Text:      text,
		CreatedAt: time.Now(),
	}
	return s.repo.Create(ctx, comment)
}

func (s *commentService) Delete(ctx context.Context, id uuid.UUID, authorId uuid.UUID) error {
	c, err := s.repo.GetById(ctx, id)
	if err != nil {
		return errors.New("comment not found")
	}
	if c.AuthorId != authorId {
		return errors.New("forbidden")
	}
	return s.repo.Delete(ctx, id)
}

func (s *commentService) GetByContentId(ctx context.Context, contentID uuid.UUID) ([]Comment, error) {
	return s.repo.GetByContentId(ctx, contentID)
}
