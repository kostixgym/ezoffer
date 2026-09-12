package ezoffer

import (
	"context"

	"ezoffer/pkg/db"
)

type InterviewListParams struct {
	Search    *string
	Grades    []string
	Types     []string
	CompanyID *int64
	IsReal    *bool
	Page      int
	PageSize  int
}


func (m *Manager) Interviews(ctx context.Context, p InterviewListParams) ([]db.Interview, int, error) {
	visible := true
	search := &db.InterviewSearch{IsVisible: &visible}

	if text := strValue(p.Search); text != "" {
		search.TitleILike = &text
	}

	grades, err := canonicalAll("grades", p.Grades, gradeValues)
	if err != nil {
		return nil, 0, err
	}

	if len(grades) > 0 {
		search.GradesIntersect = grades
	}

	types, err := canonicalAll("types", p.Types, interviewTypeValues)
	if err != nil {
		return nil, 0, err
	}

	if len(types) > 0 {
		search.TypesIntersect = types
	}

	if p.CompanyID != nil {
		search.CompanyID = p.CompanyID
	}

	if p.IsReal != nil {
		search.IsReal = p.IsReal
	}

	listOps := []db.OpFunc{
		m.repo.FullInterview(),
		db.WithSort(
			db.SortField{Column: db.Columns.Interview.PublishedDate, Direction: db.SortDescNullsLast},
			db.SortField{Column: db.Columns.Interview.ID, Direction: db.SortDesc},
		),
	}

	interviews, err := m.repo.InterviewsByFilters(ctx, search, db.NewPager(p.Page, p.PageSize), listOps...)
	if err != nil {
		return nil, 0, err
	}

	count, err := m.repo.CountInterviews(ctx, search)
	if err != nil {
		return nil, 0, err
	}

	return interviews, count, nil
}

func (m *Manager) Interview(ctx context.Context, id int64) (*db.Interview, error) {
	interview, err := m.repo.InterviewByID(ctx, id, m.repo.FullInterview())
	if err != nil {
		return nil, err
	}

	if interview == nil || !interview.IsVisible {
		return nil, nil
	}

	return interview, nil
}
