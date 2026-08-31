package rpc

import (
	"context"
	"time"

	"ezoffer/pkg/ezoffer"

	"github.com/vmkteam/embedlog"
	"github.com/vmkteam/zenrpc/v2"
)

// TestAssignmentSummary is a list entry: the statement is replaced by a short excerpt.
type TestAssignmentSummary struct {
	ID            int64      `json:"id"`
	Slug          string     `json:"slug"`
	Title         string     `json:"title"`
	Grades        []string   `json:"grades"`
	Companies     []Company  `json:"companies"`
	Skills        []Skill    `json:"skills"`
	Excerpt       string     `json:"excerpt"`
	PublishedDate *time.Time `json:"publishedDate"`
}

// TestAssignment is the full card, with the statement and the source link.
type TestAssignment struct {
	ID            int64      `json:"id"`
	Slug          string     `json:"slug"`
	Title         string     `json:"title"`
	Grades        []string   `json:"grades"`
	Companies     []Company  `json:"companies"`
	Skills        []Skill    `json:"skills"`
	Content       string     `json:"content"`
	SourceUrl     *string    `json:"sourceUrl"`
	PublishedDate *time.Time `json:"publishedDate"`
}

type TestAssignmentList struct {
	Items      []TestAssignmentSummary `json:"items"`
	TotalCount int                     `json:"totalCount"`
}

func NewTestAssignmentSummary(in ezoffer.TestAssignmentItem) TestAssignmentSummary {
	return TestAssignmentSummary{
		ID:            in.TestAssignment.ID,
		Slug:          in.TestAssignment.Slug,
		Title:         in.TestAssignment.Title,
		Grades:        stringSlice(in.TestAssignment.Grades),
		Companies:     NewCompanies(in.Companies),
		Skills:        NewSkills(in.Skills),
		Excerpt:       excerpt(in.TestAssignment.Content),
		PublishedDate: in.TestAssignment.PublishedDate,
	}
}

func NewTestAssignment(in ezoffer.TestAssignmentItem) TestAssignment {
	return TestAssignment{
		ID:            in.TestAssignment.ID,
		Slug:          in.TestAssignment.Slug,
		Title:         in.TestAssignment.Title,
		Grades:        stringSlice(in.TestAssignment.Grades),
		Companies:     NewCompanies(in.Companies),
		Skills:        NewSkills(in.Skills),
		Content:       in.TestAssignment.Content,
		SourceUrl:     in.TestAssignment.SourceUrl,
		PublishedDate: in.TestAssignment.PublishedDate,
	}
}

type TestAssignmentService struct {
	zenrpc.Service
	embedlog.Logger

	m *ezoffer.Manager
}

func NewTestAssignmentService(m *ezoffer.Manager, logger embedlog.Logger) *TestAssignmentService {
	return &TestAssignmentService{Logger: logger, m: m}
}

// List returns a page of test assignments, newest published first.
//
//zenrpc:page=1 page number, starts at 1
//zenrpc:pageSize=25 items per page, capped at 100
//zenrpc:search substring to search in title
//zenrpc:grades filter by grades: junior, middle, senior, lead
//zenrpc:skillIDs keep assignments requiring at least one of these skills
//zenrpc:companyID filter by company
//zenrpc:return test assignments page with total count
//zenrpc:400 unknown grade
//zenrpc:500 internal error
func (s TestAssignmentService) List(ctx context.Context, page, pageSize int, search *string, grades []string, skillIDs []int64, companyID *int64) (*TestAssignmentList, error) {
	page, pageSize = normalizePager(page, pageSize)

	assignments, totalCount, err := s.m.TestAssignments(ctx, ezoffer.TestAssignmentListParams{
		Search:    search,
		Grades:    grades,
		SkillIDs:  skillIDs,
		CompanyID: companyID,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, listError(ctx, s.Logger, "failed to list test assignments", err)
	}

	items := make([]TestAssignmentSummary, 0, len(assignments))
	for _, a := range assignments {
		items = append(items, NewTestAssignmentSummary(a))
	}

	return &TestAssignmentList{Items: items, TotalCount: totalCount}, nil
}

// Get returns a single test assignment by id.
//
//zenrpc:id test assignment id
//zenrpc:return test assignment
//zenrpc:404 test assignment not found
//zenrpc:500 internal error
func (s TestAssignmentService) Get(ctx context.Context, id int64) (*TestAssignment, error) {
	assignment, err := s.m.TestAssignment(ctx, id)
	if err != nil {
		s.Error(ctx, "failed to get test assignment", "err", err, "id", id)
		return nil, ErrInternal
	}

	if assignment == nil {
		return nil, ErrNotFound
	}

	out := NewTestAssignment(*assignment)

	return &out, nil
}
