package db

import (
	"github.com/go-pg/pg/v10/orm"
)

// WithQuestionsGradeContentILike filters a questionsGrades query by the text of
// the question each row belongs to. It is bound to that query: the condition is
// written against the "t" alias of questionsGrades.
//
// A semi-join rather than a condition on the joined relation, because
// CountQuestionsGrades runs without the join and would break on the relation
// alias, leaving totalCount silently out of sync with the page.
func WithQuestionsGradeContentILike(v string) OpFunc {
	return func(query *orm.Query) {
		query.Where(
			`"t"."questionId" IN (SELECT "questionId" FROM "questions" WHERE "content" ILIKE ?)`,
			"%"+v+"%",
		)
	}
}

// WithTaskCompanyID filters a tasks query by company. Bound to that query: the
// condition is written against the "t" alias of tasks.
//
// A semi-join for the same reason as above — CountTasks runs without any join.
func WithTaskCompanyID(companyID int64) OpFunc {
	return func(query *orm.Query) {
		query.Where(
			`"t"."taskId" IN (SELECT "taskId" FROM "tasksCompanies" WHERE "companyId" = ?)`,
			companyID,
		)
	}
}
