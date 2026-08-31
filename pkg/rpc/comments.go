package rpc

import (
	"context"
	"time"

	"ezoffer/pkg/db"
	"ezoffer/pkg/ezoffer"

	"github.com/vmkteam/embedlog"
	"github.com/vmkteam/zenrpc/v2"
)

type Comment struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	LikesCnt  int64     `json:"likesCnt"`
	CreatedAt time.Time `json:"createdAt"`
}

type CommentList struct {
	Items      []Comment `json:"items"`
	TotalCount int       `json:"totalCount"`
}

func NewComment(in db.Comment) Comment {
	return Comment{
		ID:        in.ID,
		Content:   in.Content,
		LikesCnt:  in.LikesCnt,
		CreatedAt: in.CreatedAt,
	}
}

type CommentService struct {
	zenrpc.Service
	embedlog.Logger

	m *ezoffer.Manager
}

func NewCommentService(m *ezoffer.Manager, logger embedlog.Logger) *CommentService {
	return &CommentService{Logger: logger, m: m}
}

// List returns a page of comments on one entity, newest first.
//
//zenrpc:entity question, task, testAssignment or interview
//zenrpc:entityID id of that entity
//zenrpc:page=1 page number, starts at 1
//zenrpc:pageSize=25 items per page, capped at 100
//zenrpc:byLikes=false order by likes instead of date
//zenrpc:return comments page with total count
//zenrpc:400 unknown entity
//zenrpc:500 internal error
func (s CommentService) List(ctx context.Context, entity string, entityID int64, page, pageSize int, byLikes bool) (*CommentList, error) {
	page, pageSize = normalizePager(page, pageSize)

	comments, totalCount, err := s.m.Comments(ctx, ezoffer.CommentListParams{
		Entity:   entity,
		EntityID: entityID,
		ByLikes:  byLikes,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, listError(ctx, s.Logger, "failed to list comments", err)
	}

	items := make([]Comment, 0, len(comments))
	for _, c := range comments {
		items = append(items, NewComment(c))
	}

	return &CommentList{Items: items, TotalCount: totalCount}, nil
}

// Create adds an anonymous comment to an entity.
//
//zenrpc:entity question, task, testAssignment or interview
//zenrpc:entityID id of that entity
//zenrpc:content comment text, up to 4000 characters
//zenrpc:return created comment
//zenrpc:400 unknown entity or empty content
//zenrpc:404 entity not found
//zenrpc:500 internal error
func (s CommentService) Create(ctx context.Context, entity string, entityID int64, content string) (*Comment, error) {
	comment, err := s.m.AddComment(ctx, entity, entityID, content)
	if err != nil {
		return nil, listError(ctx, s.Logger, "failed to add comment", err)
	}

	if comment == nil {
		return nil, ErrNotFound
	}

	out := NewComment(*comment)

	return &out, nil
}

// Like adds one like to a comment and returns the new counter.
//
//zenrpc:id comment id
//zenrpc:return new like count
//zenrpc:404 comment not found
//zenrpc:500 internal error
func (s CommentService) Like(ctx context.Context, id int64) (*int64, error) {
	likesCnt, err := s.m.LikeComment(ctx, id)
	if err != nil {
		s.Error(ctx, "failed to like comment", "err", err, "id", id)
		return nil, ErrInternal
	}

	if likesCnt == nil {
		return nil, ErrNotFound
	}

	return likesCnt, nil
}
