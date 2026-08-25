package ezoffer

import (
	"context"

	"ezoffer/pkg/db"
)

type QuestionListParams struct {
	Search   *string
	Page     int
	PageSize int
}

func (m *Manager) Questions(ctx context.Context, p QuestionListParams) ([]db.Question, int, error) {
	search := &db.QuestionSearch{}
	if p.Search != nil && *p.Search != "" {
		search.ContentILike = p.Search
	}

	pager := db.NewPager(p.Page, p.PageSize)
	sort := db.WithSort(db.SortField{
		Column:    db.Columns.Question.Frequency,
		Direction: db.SortDescNullsLast,
	})

	questions, err := m.repo.QuestionsByFilters(ctx, search, pager, sort)
	if err != nil {
		return nil, 0, err
	}

	count, err := m.repo.CountQuestions(ctx, search)
	if err != nil {
		return nil, 0, err
	}

	return questions, count, nil
}

func (m *Manager) Question(ctx context.Context, id int64) (*db.Question, error) {
	return m.repo.QuestionByID(ctx, id)
}
