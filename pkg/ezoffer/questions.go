package ezoffer

import (
	"context"

	"ezoffer/pkg/db"
)

type QuestionListParams struct {
	Search   *string
	Grade    *string
	Page     int
	PageSize int
}

// QuestionItem is a question plus, when filtered by grade, the chance for it.
type QuestionItem struct {
	Question    db.Question
	GradeChance *float32
}

func (m *Manager) Questions(ctx context.Context, p QuestionListParams) ([]QuestionItem, int, error) {
	if raw := strValue(p.Grade); raw != "" {
		grade, err := canonical("grade", raw, gradeValues)
		if err != nil {
			return nil, 0, err
		}

		return m.questionsByGrade(ctx, p, grade)
	}

	return m.questionsAll(ctx, p)
}

func (m *Manager) questionsAll(ctx context.Context, p QuestionListParams) ([]QuestionItem, int, error) {
	search := &db.QuestionSearch{}
	if text := strValue(p.Search); text != "" {
		search.ContentILike = &text
	}

	pager := db.NewPager(p.Page, p.PageSize)

	sort := db.WithSort(
		db.SortField{Column: db.Columns.Question.Frequency, Direction: db.SortDescNullsLast},
		db.SortField{Column: db.Columns.Question.ID, Direction: db.SortAsc},
	)

	questions, err := m.repo.QuestionsByFilters(ctx, search, pager, sort)
	if err != nil {
		return nil, 0, err
	}

	count, err := m.repo.CountQuestions(ctx, search)
	if err != nil {
		return nil, 0, err
	}

	items := make([]QuestionItem, 0, len(questions))
	for _, q := range questions {
		items = append(items, QuestionItem{Question: q})
	}

	return items, count, nil
}

func (m *Manager) questionsByGrade(ctx context.Context, p QuestionListParams, grade string) ([]QuestionItem, int, error) {
	search := &db.QuestionsGradeSearch{Grade: &grade}
	pager := db.NewPager(p.Page, p.PageSize)

	var ops []db.OpFunc
	if text := strValue(p.Search); text != "" {
		ops = append(ops, db.WithQuestionsGradeContentILike(text))
	}

	listOps := append([]db.OpFunc{
		m.repo.FullQuestionsGrade(),
		db.WithSort(
			db.SortField{Column: db.Columns.QuestionsGrade.Frequency, Direction: db.SortDescNullsLast},
			db.SortField{Column: db.Columns.QuestionsGrade.QuestionID, Direction: db.SortAsc},
		),
	}, ops...)

	grades, err := m.repo.QuestionsGradesByFilters(ctx, search, pager, listOps...)
	if err != nil {
		return nil, 0, err
	}

	count, err := m.repo.CountQuestionsGrades(ctx, search, ops...)
	if err != nil {
		return nil, 0, err
	}

	items := make([]QuestionItem, 0, len(grades))
	for _, g := range grades {
		if g.Question == nil {
			m.Error(ctx, "questionsGrade without a question", "questionId", g.QuestionID, "grade", g.Grade)

			continue
		}

		items = append(items, QuestionItem{Question: *g.Question, GradeChance: &g.Frequency})
	}

	return items, count, nil
}

func (m *Manager) Question(ctx context.Context, id int64) (*db.Question, error) {
	return m.repo.QuestionByID(ctx, id)
}
