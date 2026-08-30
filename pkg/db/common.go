package db

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type CommonRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewCommonRepo returns new repository
func NewCommonRepo(db orm.DB) CommonRepo {
	return CommonRepo{
		db:      db,
		filters: map[string][]Filter{},
		sort: map[string][]SortField{
			Tables.User.Name:                   {{Column: Columns.User.CreatedAt, Direction: SortDesc}},
			Tables.Comment.Name:                {{Column: Columns.Comment.CreatedAt, Direction: SortDesc}},
			Tables.Company.Name:                {{Column: Columns.Company.ID, Direction: SortDesc}},
			Tables.Interview.Name:              {{Column: Columns.Interview.CreatedAt, Direction: SortDesc}},
			Tables.Question.Name:               {{Column: Columns.Question.CreatedAt, Direction: SortDesc}},
			Tables.QuestionsGrade.Name:         {{Column: Columns.QuestionsGrade.QuestionID, Direction: SortDesc}},
			Tables.Skill.Name:                  {{Column: Columns.Skill.ID, Direction: SortDesc}},
			Tables.Task.Name:                   {{Column: Columns.Task.CreatedAt, Direction: SortDesc}},
			Tables.TasksCompany.Name:           {{Column: Columns.TasksCompany.TaskID, Direction: SortDesc}},
			Tables.TestAssignment.Name:         {{Column: Columns.TestAssignment.CreatedAt, Direction: SortDesc}},
			Tables.TestAssignmentsCompany.Name: {{Column: Columns.TestAssignmentsCompany.TestAssignmentID, Direction: SortDesc}},
			Tables.TestAssignmentsSkill.Name:   {{Column: Columns.TestAssignmentsSkill.TestAssignmentID, Direction: SortDesc}},
		},
		join: map[string][]string{
			Tables.User.Name:                   {TableColumns},
			Tables.Comment.Name:                {TableColumns, Columns.Comment.User, Columns.Comment.Interview, Columns.Comment.Question, Columns.Comment.TestAssignment, Columns.Comment.Task},
			Tables.Company.Name:                {TableColumns},
			Tables.Interview.Name:              {TableColumns, Columns.Interview.Company},
			Tables.Question.Name:               {TableColumns},
			Tables.QuestionsGrade.Name:         {TableColumns, Columns.QuestionsGrade.Question},
			Tables.Skill.Name:                  {TableColumns},
			Tables.Task.Name:                   {TableColumns},
			Tables.TasksCompany.Name:           {TableColumns, Columns.TasksCompany.Task, Columns.TasksCompany.Company},
			Tables.TestAssignment.Name:         {TableColumns},
			Tables.TestAssignmentsCompany.Name: {TableColumns, Columns.TestAssignmentsCompany.TestAssignment, Columns.TestAssignmentsCompany.Company},
			Tables.TestAssignmentsSkill.Name:   {TableColumns, Columns.TestAssignmentsSkill.TestAssignment, Columns.TestAssignmentsSkill.Skill},
		},
	}
}

// WithTransaction is a function that wraps CommonRepo with pg.Tx transaction.
func (cr CommonRepo) WithTransaction(tx *pg.Tx) CommonRepo {
	cr.db = tx
	return cr
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func (cr CommonRepo) WithEnabledOnly() CommonRepo {
	f := make(map[string][]Filter, len(cr.filters))
	for i := range cr.filters {
		f[i] = make([]Filter, len(cr.filters[i]))
		copy(f[i], cr.filters[i])
		f[i] = append(f[i], StatusEnabledFilter)
	}
	cr.filters = f

	return cr
}

/*** User ***/

// FullUser returns full joins with all columns
func (cr CommonRepo) FullUser() OpFunc {
	return WithColumns(cr.join[Tables.User.Name]...)
}

// DefaultUserSort returns default sort.
func (cr CommonRepo) DefaultUserSort() OpFunc {
	return WithSort(cr.sort[Tables.User.Name]...)
}

// UserByID is a function that returns User by ID(s) or nil.
func (cr CommonRepo) UserByID(ctx context.Context, id int64, ops ...OpFunc) (*User, error) {
	return cr.OneUser(ctx, &UserSearch{ID: &id}, ops...)
}

// OneUser is a function that returns one User by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneUser(ctx context.Context, search *UserSearch, ops ...OpFunc) (*User, error) {
	obj := &User{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.User.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// UsersByFilters returns User list.
func (cr CommonRepo) UsersByFilters(ctx context.Context, search *UserSearch, pager Pager, ops ...OpFunc) (users []User, err error) {
	err = buildQuery(ctx, cr.db, &users, search, cr.filters[Tables.User.Name], pager, ops...).Select()
	return
}

// CountUsers returns count
func (cr CommonRepo) CountUsers(ctx context.Context, search *UserSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &User{}, search, cr.filters[Tables.User.Name], PagerOne, ops...).Count()
}

// AddUser adds User to DB.
func (cr CommonRepo) AddUser(ctx context.Context, user *User, ops ...OpFunc) (*User, error) {
	q := cr.db.ModelContext(ctx, user)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.User.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return user, err
}

// UpdateUser updates User in DB.
func (cr CommonRepo) UpdateUser(ctx context.Context, user *User, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, user).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.User.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteUser deletes User from DB.
func (cr CommonRepo) DeleteUser(ctx context.Context, id int64) (deleted bool, err error) {
	user := &User{ID: id}

	res, err := cr.db.ModelContext(ctx, user).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** Comment ***/

// FullComment returns full joins with all columns
func (cr CommonRepo) FullComment() OpFunc {
	return WithColumns(cr.join[Tables.Comment.Name]...)
}

// DefaultCommentSort returns default sort.
func (cr CommonRepo) DefaultCommentSort() OpFunc {
	return WithSort(cr.sort[Tables.Comment.Name]...)
}

// CommentByID is a function that returns Comment by ID(s) or nil.
func (cr CommonRepo) CommentByID(ctx context.Context, id int64, ops ...OpFunc) (*Comment, error) {
	return cr.OneComment(ctx, &CommentSearch{ID: &id}, ops...)
}

// OneComment is a function that returns one Comment by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneComment(ctx context.Context, search *CommentSearch, ops ...OpFunc) (*Comment, error) {
	obj := &Comment{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.Comment.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// CommentsByFilters returns Comment list.
func (cr CommonRepo) CommentsByFilters(ctx context.Context, search *CommentSearch, pager Pager, ops ...OpFunc) (comments []Comment, err error) {
	err = buildQuery(ctx, cr.db, &comments, search, cr.filters[Tables.Comment.Name], pager, ops...).Select()
	return
}

// CountComments returns count
func (cr CommonRepo) CountComments(ctx context.Context, search *CommentSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &Comment{}, search, cr.filters[Tables.Comment.Name], PagerOne, ops...).Count()
}

// AddComment adds Comment to DB.
func (cr CommonRepo) AddComment(ctx context.Context, comment *Comment, ops ...OpFunc) (*Comment, error) {
	q := cr.db.ModelContext(ctx, comment)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Comment.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return comment, err
}

// UpdateComment updates Comment in DB.
func (cr CommonRepo) UpdateComment(ctx context.Context, comment *Comment, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, comment).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Comment.ID, Columns.Comment.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteComment deletes Comment from DB.
func (cr CommonRepo) DeleteComment(ctx context.Context, id int64) (deleted bool, err error) {
	comment := &Comment{ID: id}

	res, err := cr.db.ModelContext(ctx, comment).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** Company ***/

// FullCompany returns full joins with all columns
func (cr CommonRepo) FullCompany() OpFunc {
	return WithColumns(cr.join[Tables.Company.Name]...)
}

// DefaultCompanySort returns default sort.
func (cr CommonRepo) DefaultCompanySort() OpFunc {
	return WithSort(cr.sort[Tables.Company.Name]...)
}

// CompanyByID is a function that returns Company by ID(s) or nil.
func (cr CommonRepo) CompanyByID(ctx context.Context, id int64, ops ...OpFunc) (*Company, error) {
	return cr.OneCompany(ctx, &CompanySearch{ID: &id}, ops...)
}

// OneCompany is a function that returns one Company by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneCompany(ctx context.Context, search *CompanySearch, ops ...OpFunc) (*Company, error) {
	obj := &Company{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.Company.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// CompaniesByFilters returns Company list.
func (cr CommonRepo) CompaniesByFilters(ctx context.Context, search *CompanySearch, pager Pager, ops ...OpFunc) (companies []Company, err error) {
	err = buildQuery(ctx, cr.db, &companies, search, cr.filters[Tables.Company.Name], pager, ops...).Select()
	return
}

// CountCompanies returns count
func (cr CommonRepo) CountCompanies(ctx context.Context, search *CompanySearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &Company{}, search, cr.filters[Tables.Company.Name], PagerOne, ops...).Count()
}

// AddCompany adds Company to DB.
func (cr CommonRepo) AddCompany(ctx context.Context, company *Company, ops ...OpFunc) (*Company, error) {
	q := cr.db.ModelContext(ctx, company)
	applyOps(q, ops...)
	_, err := q.Insert()

	return company, err
}

// UpdateCompany updates Company in DB.
func (cr CommonRepo) UpdateCompany(ctx context.Context, company *Company, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, company).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Company.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteCompany deletes Company from DB.
func (cr CommonRepo) DeleteCompany(ctx context.Context, id int64) (deleted bool, err error) {
	company := &Company{ID: id}

	res, err := cr.db.ModelContext(ctx, company).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** Interview ***/

// FullInterview returns full joins with all columns
func (cr CommonRepo) FullInterview() OpFunc {
	return WithColumns(cr.join[Tables.Interview.Name]...)
}

// DefaultInterviewSort returns default sort.
func (cr CommonRepo) DefaultInterviewSort() OpFunc {
	return WithSort(cr.sort[Tables.Interview.Name]...)
}

// InterviewByID is a function that returns Interview by ID(s) or nil.
func (cr CommonRepo) InterviewByID(ctx context.Context, id int64, ops ...OpFunc) (*Interview, error) {
	return cr.OneInterview(ctx, &InterviewSearch{ID: &id}, ops...)
}

// OneInterview is a function that returns one Interview by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneInterview(ctx context.Context, search *InterviewSearch, ops ...OpFunc) (*Interview, error) {
	obj := &Interview{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.Interview.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// InterviewsByFilters returns Interview list.
func (cr CommonRepo) InterviewsByFilters(ctx context.Context, search *InterviewSearch, pager Pager, ops ...OpFunc) (interviews []Interview, err error) {
	err = buildQuery(ctx, cr.db, &interviews, search, cr.filters[Tables.Interview.Name], pager, ops...).Select()
	return
}

// CountInterviews returns count
func (cr CommonRepo) CountInterviews(ctx context.Context, search *InterviewSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &Interview{}, search, cr.filters[Tables.Interview.Name], PagerOne, ops...).Count()
}

// AddInterview adds Interview to DB.
func (cr CommonRepo) AddInterview(ctx context.Context, interview *Interview, ops ...OpFunc) (*Interview, error) {
	q := cr.db.ModelContext(ctx, interview)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Interview.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return interview, err
}

// UpdateInterview updates Interview in DB.
func (cr CommonRepo) UpdateInterview(ctx context.Context, interview *Interview, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, interview).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Interview.ID, Columns.Interview.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteInterview deletes Interview from DB.
func (cr CommonRepo) DeleteInterview(ctx context.Context, id int64) (deleted bool, err error) {
	interview := &Interview{ID: id}

	res, err := cr.db.ModelContext(ctx, interview).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** Question ***/

// FullQuestion returns full joins with all columns
func (cr CommonRepo) FullQuestion() OpFunc {
	return WithColumns(cr.join[Tables.Question.Name]...)
}

// DefaultQuestionSort returns default sort.
func (cr CommonRepo) DefaultQuestionSort() OpFunc {
	return WithSort(cr.sort[Tables.Question.Name]...)
}

// QuestionByID is a function that returns Question by ID(s) or nil.
func (cr CommonRepo) QuestionByID(ctx context.Context, id int64, ops ...OpFunc) (*Question, error) {
	return cr.OneQuestion(ctx, &QuestionSearch{ID: &id}, ops...)
}

// OneQuestion is a function that returns one Question by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneQuestion(ctx context.Context, search *QuestionSearch, ops ...OpFunc) (*Question, error) {
	obj := &Question{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.Question.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// QuestionsByFilters returns Question list.
func (cr CommonRepo) QuestionsByFilters(ctx context.Context, search *QuestionSearch, pager Pager, ops ...OpFunc) (questions []Question, err error) {
	err = buildQuery(ctx, cr.db, &questions, search, cr.filters[Tables.Question.Name], pager, ops...).Select()
	return
}

// CountQuestions returns count
func (cr CommonRepo) CountQuestions(ctx context.Context, search *QuestionSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &Question{}, search, cr.filters[Tables.Question.Name], PagerOne, ops...).Count()
}

// AddQuestion adds Question to DB.
func (cr CommonRepo) AddQuestion(ctx context.Context, question *Question, ops ...OpFunc) (*Question, error) {
	q := cr.db.ModelContext(ctx, question)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Question.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return question, err
}

// UpdateQuestion updates Question in DB.
func (cr CommonRepo) UpdateQuestion(ctx context.Context, question *Question, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, question).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Question.ID, Columns.Question.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteQuestion deletes Question from DB.
func (cr CommonRepo) DeleteQuestion(ctx context.Context, id int64) (deleted bool, err error) {
	question := &Question{ID: id}

	res, err := cr.db.ModelContext(ctx, question).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** QuestionsGrade ***/

// FullQuestionsGrade returns full joins with all columns
func (cr CommonRepo) FullQuestionsGrade() OpFunc {
	return WithColumns(cr.join[Tables.QuestionsGrade.Name]...)
}

// DefaultQuestionsGradeSort returns default sort.
func (cr CommonRepo) DefaultQuestionsGradeSort() OpFunc {
	return WithSort(cr.sort[Tables.QuestionsGrade.Name]...)
}

// QuestionsGradeByID is a function that returns QuestionsGrade by ID(s) or nil.
func (cr CommonRepo) QuestionsGradeByID(ctx context.Context, questionID int64, grade string, ops ...OpFunc) (*QuestionsGrade, error) {
	return cr.OneQuestionsGrade(ctx, &QuestionsGradeSearch{QuestionID: &questionID, Grade: &grade}, ops...)
}

// OneQuestionsGrade is a function that returns one QuestionsGrade by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneQuestionsGrade(ctx context.Context, search *QuestionsGradeSearch, ops ...OpFunc) (*QuestionsGrade, error) {
	obj := &QuestionsGrade{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.QuestionsGrade.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// QuestionsGradesByFilters returns QuestionsGrade list.
func (cr CommonRepo) QuestionsGradesByFilters(ctx context.Context, search *QuestionsGradeSearch, pager Pager, ops ...OpFunc) (questionsGrades []QuestionsGrade, err error) {
	err = buildQuery(ctx, cr.db, &questionsGrades, search, cr.filters[Tables.QuestionsGrade.Name], pager, ops...).Select()
	return
}

// CountQuestionsGrades returns count
func (cr CommonRepo) CountQuestionsGrades(ctx context.Context, search *QuestionsGradeSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &QuestionsGrade{}, search, cr.filters[Tables.QuestionsGrade.Name], PagerOne, ops...).Count()
}

// AddQuestionsGrade adds QuestionsGrade to DB.
func (cr CommonRepo) AddQuestionsGrade(ctx context.Context, questionsGrade *QuestionsGrade, ops ...OpFunc) (*QuestionsGrade, error) {
	q := cr.db.ModelContext(ctx, questionsGrade)
	applyOps(q, ops...)
	_, err := q.Insert()

	return questionsGrade, err
}

// UpdateQuestionsGrade updates QuestionsGrade in DB.
func (cr CommonRepo) UpdateQuestionsGrade(ctx context.Context, questionsGrade *QuestionsGrade, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, questionsGrade).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.QuestionsGrade.QuestionID, Columns.QuestionsGrade.Grade)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteQuestionsGrade deletes QuestionsGrade from DB.
func (cr CommonRepo) DeleteQuestionsGrade(ctx context.Context, questionID int64, grade string) (deleted bool, err error) {
	questionsGrade := &QuestionsGrade{QuestionID: questionID, Grade: grade}

	res, err := cr.db.ModelContext(ctx, questionsGrade).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** Skill ***/

// FullSkill returns full joins with all columns
func (cr CommonRepo) FullSkill() OpFunc {
	return WithColumns(cr.join[Tables.Skill.Name]...)
}

// DefaultSkillSort returns default sort.
func (cr CommonRepo) DefaultSkillSort() OpFunc {
	return WithSort(cr.sort[Tables.Skill.Name]...)
}

// SkillByID is a function that returns Skill by ID(s) or nil.
func (cr CommonRepo) SkillByID(ctx context.Context, id int64, ops ...OpFunc) (*Skill, error) {
	return cr.OneSkill(ctx, &SkillSearch{ID: &id}, ops...)
}

// OneSkill is a function that returns one Skill by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneSkill(ctx context.Context, search *SkillSearch, ops ...OpFunc) (*Skill, error) {
	obj := &Skill{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.Skill.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// SkillsByFilters returns Skill list.
func (cr CommonRepo) SkillsByFilters(ctx context.Context, search *SkillSearch, pager Pager, ops ...OpFunc) (skills []Skill, err error) {
	err = buildQuery(ctx, cr.db, &skills, search, cr.filters[Tables.Skill.Name], pager, ops...).Select()
	return
}

// CountSkills returns count
func (cr CommonRepo) CountSkills(ctx context.Context, search *SkillSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &Skill{}, search, cr.filters[Tables.Skill.Name], PagerOne, ops...).Count()
}

// AddSkill adds Skill to DB.
func (cr CommonRepo) AddSkill(ctx context.Context, skill *Skill, ops ...OpFunc) (*Skill, error) {
	q := cr.db.ModelContext(ctx, skill)
	applyOps(q, ops...)
	_, err := q.Insert()

	return skill, err
}

// UpdateSkill updates Skill in DB.
func (cr CommonRepo) UpdateSkill(ctx context.Context, skill *Skill, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, skill).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Skill.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteSkill deletes Skill from DB.
func (cr CommonRepo) DeleteSkill(ctx context.Context, id int64) (deleted bool, err error) {
	skill := &Skill{ID: id}

	res, err := cr.db.ModelContext(ctx, skill).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** Task ***/

// FullTask returns full joins with all columns
func (cr CommonRepo) FullTask() OpFunc {
	return WithColumns(cr.join[Tables.Task.Name]...)
}

// DefaultTaskSort returns default sort.
func (cr CommonRepo) DefaultTaskSort() OpFunc {
	return WithSort(cr.sort[Tables.Task.Name]...)
}

// TaskByID is a function that returns Task by ID(s) or nil.
func (cr CommonRepo) TaskByID(ctx context.Context, id int64, ops ...OpFunc) (*Task, error) {
	return cr.OneTask(ctx, &TaskSearch{ID: &id}, ops...)
}

// OneTask is a function that returns one Task by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneTask(ctx context.Context, search *TaskSearch, ops ...OpFunc) (*Task, error) {
	obj := &Task{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.Task.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// TasksByFilters returns Task list.
func (cr CommonRepo) TasksByFilters(ctx context.Context, search *TaskSearch, pager Pager, ops ...OpFunc) (tasks []Task, err error) {
	err = buildQuery(ctx, cr.db, &tasks, search, cr.filters[Tables.Task.Name], pager, ops...).Select()
	return
}

// CountTasks returns count
func (cr CommonRepo) CountTasks(ctx context.Context, search *TaskSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &Task{}, search, cr.filters[Tables.Task.Name], PagerOne, ops...).Count()
}

// AddTask adds Task to DB.
func (cr CommonRepo) AddTask(ctx context.Context, task *Task, ops ...OpFunc) (*Task, error) {
	q := cr.db.ModelContext(ctx, task)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Task.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return task, err
}

// UpdateTask updates Task in DB.
func (cr CommonRepo) UpdateTask(ctx context.Context, task *Task, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, task).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Task.ID, Columns.Task.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteTask deletes Task from DB.
func (cr CommonRepo) DeleteTask(ctx context.Context, id int64) (deleted bool, err error) {
	task := &Task{ID: id}

	res, err := cr.db.ModelContext(ctx, task).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** TasksCompany ***/

// FullTasksCompany returns full joins with all columns
func (cr CommonRepo) FullTasksCompany() OpFunc {
	return WithColumns(cr.join[Tables.TasksCompany.Name]...)
}

// DefaultTasksCompanySort returns default sort.
func (cr CommonRepo) DefaultTasksCompanySort() OpFunc {
	return WithSort(cr.sort[Tables.TasksCompany.Name]...)
}

// TasksCompanyByID is a function that returns TasksCompany by ID(s) or nil.
func (cr CommonRepo) TasksCompanyByID(ctx context.Context, taskID int64, companyID int64, ops ...OpFunc) (*TasksCompany, error) {
	return cr.OneTasksCompany(ctx, &TasksCompanySearch{TaskID: &taskID, CompanyID: &companyID}, ops...)
}

// OneTasksCompany is a function that returns one TasksCompany by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneTasksCompany(ctx context.Context, search *TasksCompanySearch, ops ...OpFunc) (*TasksCompany, error) {
	obj := &TasksCompany{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.TasksCompany.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// TasksCompaniesByFilters returns TasksCompany list.
func (cr CommonRepo) TasksCompaniesByFilters(ctx context.Context, search *TasksCompanySearch, pager Pager, ops ...OpFunc) (tasksCompanies []TasksCompany, err error) {
	err = buildQuery(ctx, cr.db, &tasksCompanies, search, cr.filters[Tables.TasksCompany.Name], pager, ops...).Select()
	return
}

// CountTasksCompanies returns count
func (cr CommonRepo) CountTasksCompanies(ctx context.Context, search *TasksCompanySearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &TasksCompany{}, search, cr.filters[Tables.TasksCompany.Name], PagerOne, ops...).Count()
}

// AddTasksCompany adds TasksCompany to DB.
func (cr CommonRepo) AddTasksCompany(ctx context.Context, tasksCompany *TasksCompany, ops ...OpFunc) (*TasksCompany, error) {
	q := cr.db.ModelContext(ctx, tasksCompany)
	applyOps(q, ops...)
	_, err := q.Insert()

	return tasksCompany, err
}

// UpdateTasksCompany updates TasksCompany in DB.
func (cr CommonRepo) UpdateTasksCompany(ctx context.Context, tasksCompany *TasksCompany, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, tasksCompany).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.TasksCompany.TaskID, Columns.TasksCompany.CompanyID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteTasksCompany deletes TasksCompany from DB.
func (cr CommonRepo) DeleteTasksCompany(ctx context.Context, taskID int64, companyID int64) (deleted bool, err error) {
	tasksCompany := &TasksCompany{TaskID: taskID, CompanyID: companyID}

	res, err := cr.db.ModelContext(ctx, tasksCompany).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** TestAssignment ***/

// FullTestAssignment returns full joins with all columns
func (cr CommonRepo) FullTestAssignment() OpFunc {
	return WithColumns(cr.join[Tables.TestAssignment.Name]...)
}

// DefaultTestAssignmentSort returns default sort.
func (cr CommonRepo) DefaultTestAssignmentSort() OpFunc {
	return WithSort(cr.sort[Tables.TestAssignment.Name]...)
}

// TestAssignmentByID is a function that returns TestAssignment by ID(s) or nil.
func (cr CommonRepo) TestAssignmentByID(ctx context.Context, id int64, ops ...OpFunc) (*TestAssignment, error) {
	return cr.OneTestAssignment(ctx, &TestAssignmentSearch{ID: &id}, ops...)
}

// OneTestAssignment is a function that returns one TestAssignment by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneTestAssignment(ctx context.Context, search *TestAssignmentSearch, ops ...OpFunc) (*TestAssignment, error) {
	obj := &TestAssignment{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.TestAssignment.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// TestAssignmentsByFilters returns TestAssignment list.
func (cr CommonRepo) TestAssignmentsByFilters(ctx context.Context, search *TestAssignmentSearch, pager Pager, ops ...OpFunc) (testAssignments []TestAssignment, err error) {
	err = buildQuery(ctx, cr.db, &testAssignments, search, cr.filters[Tables.TestAssignment.Name], pager, ops...).Select()
	return
}

// CountTestAssignments returns count
func (cr CommonRepo) CountTestAssignments(ctx context.Context, search *TestAssignmentSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &TestAssignment{}, search, cr.filters[Tables.TestAssignment.Name], PagerOne, ops...).Count()
}

// AddTestAssignment adds TestAssignment to DB.
func (cr CommonRepo) AddTestAssignment(ctx context.Context, testAssignment *TestAssignment, ops ...OpFunc) (*TestAssignment, error) {
	q := cr.db.ModelContext(ctx, testAssignment)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.TestAssignment.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return testAssignment, err
}

// UpdateTestAssignment updates TestAssignment in DB.
func (cr CommonRepo) UpdateTestAssignment(ctx context.Context, testAssignment *TestAssignment, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, testAssignment).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.TestAssignment.ID, Columns.TestAssignment.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteTestAssignment deletes TestAssignment from DB.
func (cr CommonRepo) DeleteTestAssignment(ctx context.Context, id int64) (deleted bool, err error) {
	testAssignment := &TestAssignment{ID: id}

	res, err := cr.db.ModelContext(ctx, testAssignment).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** TestAssignmentsCompany ***/

// FullTestAssignmentsCompany returns full joins with all columns
func (cr CommonRepo) FullTestAssignmentsCompany() OpFunc {
	return WithColumns(cr.join[Tables.TestAssignmentsCompany.Name]...)
}

// DefaultTestAssignmentsCompanySort returns default sort.
func (cr CommonRepo) DefaultTestAssignmentsCompanySort() OpFunc {
	return WithSort(cr.sort[Tables.TestAssignmentsCompany.Name]...)
}

// TestAssignmentsCompanyByID is a function that returns TestAssignmentsCompany by ID(s) or nil.
func (cr CommonRepo) TestAssignmentsCompanyByID(ctx context.Context, testAssignmentID int64, companyID int64, ops ...OpFunc) (*TestAssignmentsCompany, error) {
	return cr.OneTestAssignmentsCompany(ctx, &TestAssignmentsCompanySearch{TestAssignmentID: &testAssignmentID, CompanyID: &companyID}, ops...)
}

// OneTestAssignmentsCompany is a function that returns one TestAssignmentsCompany by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneTestAssignmentsCompany(ctx context.Context, search *TestAssignmentsCompanySearch, ops ...OpFunc) (*TestAssignmentsCompany, error) {
	obj := &TestAssignmentsCompany{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.TestAssignmentsCompany.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// TestAssignmentsCompaniesByFilters returns TestAssignmentsCompany list.
func (cr CommonRepo) TestAssignmentsCompaniesByFilters(ctx context.Context, search *TestAssignmentsCompanySearch, pager Pager, ops ...OpFunc) (testAssignmentsCompanies []TestAssignmentsCompany, err error) {
	err = buildQuery(ctx, cr.db, &testAssignmentsCompanies, search, cr.filters[Tables.TestAssignmentsCompany.Name], pager, ops...).Select()
	return
}

// CountTestAssignmentsCompanies returns count
func (cr CommonRepo) CountTestAssignmentsCompanies(ctx context.Context, search *TestAssignmentsCompanySearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &TestAssignmentsCompany{}, search, cr.filters[Tables.TestAssignmentsCompany.Name], PagerOne, ops...).Count()
}

// AddTestAssignmentsCompany adds TestAssignmentsCompany to DB.
func (cr CommonRepo) AddTestAssignmentsCompany(ctx context.Context, testAssignmentsCompany *TestAssignmentsCompany, ops ...OpFunc) (*TestAssignmentsCompany, error) {
	q := cr.db.ModelContext(ctx, testAssignmentsCompany)
	applyOps(q, ops...)
	_, err := q.Insert()

	return testAssignmentsCompany, err
}

// UpdateTestAssignmentsCompany updates TestAssignmentsCompany in DB.
func (cr CommonRepo) UpdateTestAssignmentsCompany(ctx context.Context, testAssignmentsCompany *TestAssignmentsCompany, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, testAssignmentsCompany).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.TestAssignmentsCompany.TestAssignmentID, Columns.TestAssignmentsCompany.CompanyID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteTestAssignmentsCompany deletes TestAssignmentsCompany from DB.
func (cr CommonRepo) DeleteTestAssignmentsCompany(ctx context.Context, testAssignmentID int64, companyID int64) (deleted bool, err error) {
	testAssignmentsCompany := &TestAssignmentsCompany{TestAssignmentID: testAssignmentID, CompanyID: companyID}

	res, err := cr.db.ModelContext(ctx, testAssignmentsCompany).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** TestAssignmentsSkill ***/

// FullTestAssignmentsSkill returns full joins with all columns
func (cr CommonRepo) FullTestAssignmentsSkill() OpFunc {
	return WithColumns(cr.join[Tables.TestAssignmentsSkill.Name]...)
}

// DefaultTestAssignmentsSkillSort returns default sort.
func (cr CommonRepo) DefaultTestAssignmentsSkillSort() OpFunc {
	return WithSort(cr.sort[Tables.TestAssignmentsSkill.Name]...)
}

// TestAssignmentsSkillByID is a function that returns TestAssignmentsSkill by ID(s) or nil.
func (cr CommonRepo) TestAssignmentsSkillByID(ctx context.Context, testAssignmentID int64, skillID int64, ops ...OpFunc) (*TestAssignmentsSkill, error) {
	return cr.OneTestAssignmentsSkill(ctx, &TestAssignmentsSkillSearch{TestAssignmentID: &testAssignmentID, SkillID: &skillID}, ops...)
}

// OneTestAssignmentsSkill is a function that returns one TestAssignmentsSkill by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneTestAssignmentsSkill(ctx context.Context, search *TestAssignmentsSkillSearch, ops ...OpFunc) (*TestAssignmentsSkill, error) {
	obj := &TestAssignmentsSkill{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.TestAssignmentsSkill.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// TestAssignmentsSkillsByFilters returns TestAssignmentsSkill list.
func (cr CommonRepo) TestAssignmentsSkillsByFilters(ctx context.Context, search *TestAssignmentsSkillSearch, pager Pager, ops ...OpFunc) (testAssignmentsSkills []TestAssignmentsSkill, err error) {
	err = buildQuery(ctx, cr.db, &testAssignmentsSkills, search, cr.filters[Tables.TestAssignmentsSkill.Name], pager, ops...).Select()
	return
}

// CountTestAssignmentsSkills returns count
func (cr CommonRepo) CountTestAssignmentsSkills(ctx context.Context, search *TestAssignmentsSkillSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &TestAssignmentsSkill{}, search, cr.filters[Tables.TestAssignmentsSkill.Name], PagerOne, ops...).Count()
}

// AddTestAssignmentsSkill adds TestAssignmentsSkill to DB.
func (cr CommonRepo) AddTestAssignmentsSkill(ctx context.Context, testAssignmentsSkill *TestAssignmentsSkill, ops ...OpFunc) (*TestAssignmentsSkill, error) {
	q := cr.db.ModelContext(ctx, testAssignmentsSkill)
	applyOps(q, ops...)
	_, err := q.Insert()

	return testAssignmentsSkill, err
}

// UpdateTestAssignmentsSkill updates TestAssignmentsSkill in DB.
func (cr CommonRepo) UpdateTestAssignmentsSkill(ctx context.Context, testAssignmentsSkill *TestAssignmentsSkill, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, testAssignmentsSkill).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.TestAssignmentsSkill.TestAssignmentID, Columns.TestAssignmentsSkill.SkillID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteTestAssignmentsSkill deletes TestAssignmentsSkill from DB.
func (cr CommonRepo) DeleteTestAssignmentsSkill(ctx context.Context, testAssignmentID int64, skillID int64) (deleted bool, err error) {
	testAssignmentsSkill := &TestAssignmentsSkill{TestAssignmentID: testAssignmentID, SkillID: skillID}

	res, err := cr.db.ModelContext(ctx, testAssignmentsSkill).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}
