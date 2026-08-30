package rpc

import (
	"context"

	"ezoffer/pkg/ezoffer"

	"github.com/vmkteam/embedlog"
	"github.com/vmkteam/zenrpc/v2"
)

type Question struct {
	ID      int64  `json:"id"`
	Slug    string `json:"slug"`
	Content string `json:"content"`

	// Chance is the overall probability of getting this question, 0..100.
	// Null when the source has no statistics for it.
	Chance *float32 `json:"chance"`

	// GradeChance is the probability for a single grade. Filled in only when the
	// list is filtered by grade, null otherwise.
	GradeChance *float32 `json:"gradeChance"`
}

type QuestionList struct {
	Items      []Question `json:"items"`
	TotalCount int        `json:"totalCount"`
}

func NewQuestion(in ezoffer.QuestionItem) Question {
	return Question{
		ID:          in.Question.ID,
		Slug:        in.Question.Slug,
		Content:     in.Question.Content,
		Chance:      in.Question.Frequency,
		GradeChance: in.GradeChance,
	}
}

type QuestionService struct {
	zenrpc.Service
	embedlog.Logger

	m *ezoffer.Manager
}

func NewQuestionService(m *ezoffer.Manager, logger embedlog.Logger) *QuestionService {
	return &QuestionService{Logger: logger, m: m}
}

// List returns a page of questions ordered by chance, highest first.
//
//zenrpc:page=1 page number, starts at 1
//zenrpc:pageSize=25 items per page, capped at 100
//zenrpc:search substring to search in question text
//zenrpc:grade filter by grade: junior, middle, senior or lead
//zenrpc:return questions page with total count
//zenrpc:500 internal error
func (s QuestionService) List(ctx context.Context, page, pageSize int, search, grade *string) (*QuestionList, error) {
	page, pageSize = normalizePager(page, pageSize)

	questions, totalCount, err := s.m.Questions(ctx, ezoffer.QuestionListParams{
		Search:   search,
		Grade:    grade,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		s.Error(ctx, "failed to list questions", "err", err)
		return nil, ErrInternal
	}

	items := make([]Question, 0, len(questions))
	for _, q := range questions {
		items = append(items, NewQuestion(q))
	}

	return &QuestionList{Items: items, TotalCount: totalCount}, nil
}

// Get returns a single question by id.
//
//zenrpc:id question id
//zenrpc:return question
//zenrpc:404 question not found
//zenrpc:500 internal error
func (s QuestionService) Get(ctx context.Context, id int64) (*Question, error) {
	question, err := s.m.Question(ctx, id)
	if err != nil {
		s.Error(ctx, "failed to get question", "err", err, "id", id)
		return nil, ErrInternal
	}

	if question == nil {
		return nil, ErrNotFound
	}

	out := NewQuestion(ezoffer.QuestionItem{Question: *question})

	return &out, nil
}
