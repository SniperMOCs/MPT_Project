package comment

import (
	"context"
	"errors"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/content"
	"github.com/SniperMOCs/MPT_Project/internal/notification"
	"github.com/google/uuid"
)

type commentService struct {
	repo CommentRepository
	ns   notification.NotificationService
	cs   content.ContentService
}

func NewCommentService(repo CommentRepository, ns notification.NotificationService, cs content.ContentService) CommentService {
	return &commentService{
		repo: repo,
		ns:   ns,
		cs:   cs,
	}
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
	if err := s.repo.Create(ctx, comment); err != nil {
		return err
	}

	c, err := s.cs.GetByID(ctx, contentId)
	if err != nil {
		return err
	}

	flag := true
	if comment.ParentId != nil {
		parentComment, err := s.repo.GetById(ctx, *comment.ParentId)
		if err != nil {
			return err
		}
		if parentComment.AuthorId == c.Author.ID {
			flag = false
		}
		s.ns.Notify(ctx, authorId, parentComment.AuthorId, c.ID, notification.Response)
	}

	if flag {
		s.ns.Notify(ctx, authorId, c.Author.ID, c.ID, notification.Comment)
	}
	return nil
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

func (s *commentService) GetByContentId(ctx context.Context, contentID uuid.UUID) ([]CommentResponse, error) {
	return s.repo.GetByContentId(ctx, contentID)
}
