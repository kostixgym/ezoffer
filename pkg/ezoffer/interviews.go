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

// Interviews returns a page of interview recordings and the total count, newest
// published first.
//
// 83 of the 180 records carry neither grade nor type — those are the public ones
// scraped without that metadata. They show up in an unfiltered list and drop out
// as soon as a grade or type filter is set, because an empty array intersects
// nothing.
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

	// publishedDate is null on a good third of the rows, so id carries the order
	// there. Without it the page boundary is undefined and rows drift.
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

// Interview returns interview by id. Returns nil when it is not found or hidden.
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
