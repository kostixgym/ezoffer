package ezoffer

import (
	"context"
	"strings"

	"ezoffer/pkg/db"
)

// The four entities a comment can hang on. The table keeps one nullable foreign
// key per entity with a CHECK that exactly one of them is set, and the API
// mirrors that shape: an entity name plus an id.
const (
	EntityQuestion       = "question"
	EntityTask           = "task"
	EntityTestAssignment = "testAssignment"
	EntityInterview      = "interview"
)

var entityValues = canonicalMap(EntityQuestion, EntityTask, EntityTestAssignment, EntityInterview)

// commentMaxRunes keeps a single comment from becoming a denial of service on
// its own. Nothing in the UI needs more.
const commentMaxRunes = 4000

type CommentListParams struct {
	Entity   string
	EntityID int64
	ByLikes  bool
	Page     int
	PageSize int
}

// Comments returns a page of visible comments for one entity, newest first
// unless ordering by likes was asked for.
func (m *Manager) Comments(ctx context.Context, p CommentListParams) ([]db.Comment, int, error) {
	search, err := commentSearch(p.Entity, p.EntityID)
	if err != nil {
		return nil, 0, err
	}

	sortColumn := db.Columns.Comment.CreatedAt
	if p.ByLikes {
		sortColumn = db.Columns.Comment.LikesCnt
	}

	// createdAt is not unique — a burst of comments shares a timestamp down to
	// the microsecond only rarely, but likesCnt collides constantly, most of it
	// on zero. id keeps the page boundary stable either way.
	sort := db.WithSort(
		db.SortField{Column: sortColumn, Direction: db.SortDesc},
		db.SortField{Column: db.Columns.Comment.ID, Direction: db.SortDesc},
	)

	comments, err := m.repo.CommentsByFilters(ctx, search, db.NewPager(p.Page, p.PageSize), sort)
	if err != nil {
		return nil, 0, err
	}

	count, err := m.repo.CountComments(ctx, search)
	if err != nil {
		return nil, 0, err
	}

	return comments, count, nil
}

// AddComment attaches an anonymous comment to an entity. Returns nil when the
// entity does not exist.
func (m *Manager) AddComment(ctx context.Context, entity string, entityID int64, content string) (*db.Comment, error) {
	name, err := canonical("entity", entity, entityValues)
	if err != nil {
		return nil, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ValidationError{Field: "content", Value: content}
	}

	if len([]rune(content)) > commentMaxRunes {
		return nil, ValidationError{Field: "content", Value: "longer than 4000 characters"}
	}

	// Checked before the insert so a missing entity comes back as "not found"
	// rather than a foreign key violation the client cannot read.
	exists, err := m.entityExists(ctx, name, entityID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	comment := &db.Comment{Content: content, IsVisible: true}
	switch name {
	case EntityQuestion:
		comment.QuestionID = &entityID
	case EntityTask:
		comment.TaskID = &entityID
	case EntityTestAssignment:
		comment.TestAssignmentID = &entityID
	case EntityInterview:
		comment.InterviewID = &entityID
	}

	return m.repo.AddComment(ctx, comment)
}

// LikeComment bumps the like counter and returns the new value. Returns nil when
// there is no visible comment with that id.
//
// Deliberately not idempotent: who liked what is not stored, so a client can
// like the same comment twice. That was the accepted trade for keeping the
// comments anonymous.
func (m *Manager) LikeComment(ctx context.Context, id int64) (*int64, error) {
	return m.repo.IncCommentLikes(ctx, id)
}

func commentSearch(entity string, entityID int64) (*db.CommentSearch, error) {
	name, err := canonical("entity", entity, entityValues)
	if err != nil {
		return nil, err
	}

	visible := true
	search := &db.CommentSearch{IsVisible: &visible}

	switch name {
	case EntityQuestion:
		search.QuestionID = &entityID
	case EntityTask:
		search.TaskID = &entityID
	case EntityTestAssignment:
		search.TestAssignmentID = &entityID
	case EntityInterview:
		search.InterviewID = &entityID
	}

	return search, nil
}

func (m *Manager) entityExists(ctx context.Context, entity string, id int64) (bool, error) {
	switch entity {
	case EntityQuestion:
		v, err := m.repo.QuestionByID(ctx, id)
		return v != nil, err
	case EntityTask:
		v, err := m.repo.TaskByID(ctx, id)
		return v != nil, err
	case EntityTestAssignment:
		v, err := m.repo.TestAssignmentByID(ctx, id)
		return v != nil, err
	case EntityInterview:
		v, err := m.repo.InterviewByID(ctx, id)
		return v != nil, err
	}

	return false, nil
}
