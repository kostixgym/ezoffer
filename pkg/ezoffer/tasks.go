package ezoffer

import (
	"context"

	"ezoffer/pkg/db"
)

type TaskListParams struct {
	Search    *string
	Grades    []string
	Type      *string
	CompanyID *int64
	Page      int
	PageSize  int
}

type TaskItem struct {
	Task      db.Task
	Companies []db.Company
}


func (m *Manager) Tasks(ctx context.Context, p TaskListParams) ([]TaskItem, int, error) {
	search := &db.TaskSearch{}
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

	if raw := strValue(p.Type); raw != "" {
		t, cerr := canonical("type", raw, taskTypeValues)
		if cerr != nil {
			return nil, 0, cerr
		}

		search.Type = &t
	}

	// Company lives in a join table, so it cannot go into the generated search.
	// The op has to reach the count query as well, otherwise totalCount and items
	// stop matching.
	var ops []db.OpFunc
	if p.CompanyID != nil {
		ops = append(ops, db.WithTaskCompanyID(*p.CompanyID))
	}

	listOps := append([]db.OpFunc{
		db.WithSort(
			db.SortField{Column: db.Columns.Task.Rank, Direction: db.SortAscNullsLast},
			db.SortField{Column: db.Columns.Task.ID, Direction: db.SortAsc},
		),
	}, ops...)

	tasks, err := m.repo.TasksByFilters(ctx, search, db.NewPager(p.Page, p.PageSize), listOps...)
	if err != nil {
		return nil, 0, err
	}

	count, err := m.repo.CountTasks(ctx, search, ops...)
	if err != nil {
		return nil, 0, err
	}

	ids := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.ID)
	}

	companies, err := m.taskCompanies(ctx, ids)
	if err != nil {
		return nil, 0, err
	}

	items := make([]TaskItem, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, TaskItem{Task: t, Companies: companies[t.ID]})
	}

	return items, count, nil
}

// Task returns task by id. Returns nil when task is not found.
func (m *Manager) Task(ctx context.Context, id int64) (*TaskItem, error) {
	task, err := m.repo.TaskByID(ctx, id)
	if err != nil || task == nil {
		return nil, err
	}

	companies, err := m.taskCompanies(ctx, []int64{task.ID})
	if err != nil {
		return nil, err
	}

	return &TaskItem{Task: *task, Companies: companies[task.ID]}, nil
}

// taskCompanies loads companies for a whole page at once instead of querying per
// task. mfd generates has-one relations only from the side that holds the key,
// so tasks have no companies relation to eager load.
func (m *Manager) taskCompanies(ctx context.Context, taskIDs []int64) (map[int64][]db.Company, error) {
	out := map[int64][]db.Company{}
	if len(taskIDs) == 0 {
		return out, nil
	}

	links, err := m.repo.TasksCompaniesByFilters(ctx,
		&db.TasksCompanySearch{TaskIDs: taskIDs},
		db.PagerNoLimit,
		db.WithRelations(db.Columns.TasksCompany.Company),
	)
	if err != nil {
		return nil, err
	}

	for _, l := range links {
		if l.Company == nil {
			m.Error(ctx, "tasksCompany without a company", "taskId", l.TaskID, "companyId", l.CompanyID)

			continue
		}

		out[l.TaskID] = append(out[l.TaskID], *l.Company)
	}

	return out, nil
}
