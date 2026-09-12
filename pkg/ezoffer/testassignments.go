package ezoffer

import (
	"context"

	"ezoffer/pkg/db"
)

type TestAssignmentListParams struct {
	Search    *string
	Grades    []string
	SkillIDs  []int64
	CompanyID *int64
	Page      int
	PageSize  int
}


type TestAssignmentItem struct {
	TestAssignment db.TestAssignment
	Companies      []db.Company
	Skills         []db.Skill
}


func (m *Manager) TestAssignments(ctx context.Context, p TestAssignmentListParams) ([]TestAssignmentItem, int, error) {
	search := &db.TestAssignmentSearch{}
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

	var ops []db.OpFunc
	if p.CompanyID != nil {
		ops = append(ops, db.WithTestAssignmentCompanyID(*p.CompanyID))
	}

	if len(p.SkillIDs) > 0 {
		ops = append(ops, db.WithTestAssignmentSkillIDs(p.SkillIDs))
	}

	listOps := append([]db.OpFunc{
		db.WithSort(
			db.SortField{Column: db.Columns.TestAssignment.PublishedDate, Direction: db.SortDescNullsLast},
			db.SortField{Column: db.Columns.TestAssignment.ID, Direction: db.SortAsc},
		),
	}, ops...)

	assignments, err := m.repo.TestAssignmentsByFilters(ctx, search, db.NewPager(p.Page, p.PageSize), listOps...)
	if err != nil {
		return nil, 0, err
	}

	count, err := m.repo.CountTestAssignments(ctx, search, ops...)
	if err != nil {
		return nil, 0, err
	}

	items, err := m.testAssignmentItems(ctx, assignments)
	if err != nil {
		return nil, 0, err
	}

	return items, count, nil
}


func (m *Manager) TestAssignment(ctx context.Context, id int64) (*TestAssignmentItem, error) {
	assignment, err := m.repo.TestAssignmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if assignment == nil {
		return nil, nil
	}

	items, err := m.testAssignmentItems(ctx, []db.TestAssignment{*assignment})
	if err != nil {
		return nil, err
	}

	return &items[0], nil
}


func (m *Manager) testAssignmentItems(ctx context.Context, assignments []db.TestAssignment) ([]TestAssignmentItem, error) {
	ids := make([]int64, 0, len(assignments))
	for _, a := range assignments {
		ids = append(ids, a.ID)
	}

	companies, err := m.testAssignmentCompanies(ctx, ids)
	if err != nil {
		return nil, err
	}

	skills, err := m.testAssignmentSkills(ctx, ids)
	if err != nil {
		return nil, err
	}

	items := make([]TestAssignmentItem, 0, len(assignments))
	for _, a := range assignments {
		items = append(items, TestAssignmentItem{
			TestAssignment: a,
			Companies:      companies[a.ID],
			Skills:         skills[a.ID],
		})
	}

	return items, nil
}

func (m *Manager) testAssignmentCompanies(ctx context.Context, ids []int64) (map[int64][]db.Company, error) {
	out := map[int64][]db.Company{}
	if len(ids) == 0 {
		return out, nil
	}

	links, err := m.repo.TestAssignmentsCompaniesByFilters(ctx,
		&db.TestAssignmentsCompanySearch{TestAssignmentIDs: ids},
		db.PagerNoLimit,
		db.WithRelations(db.Columns.TestAssignmentsCompany.Company),
		db.WithSort(db.SortField{Column: db.Columns.TestAssignmentsCompany.CompanyID, Direction: db.SortAsc}),
	)
	if err != nil {
		return nil, err
	}

	for _, l := range links {
		if l.Company == nil {
			m.Error(ctx, "testAssignmentsCompany without a company", "testAssignmentId", l.TestAssignmentID, "companyId", l.CompanyID)

			continue
		}

		out[l.TestAssignmentID] = append(out[l.TestAssignmentID], *l.Company)
	}

	return out, nil
}

func (m *Manager) testAssignmentSkills(ctx context.Context, ids []int64) (map[int64][]db.Skill, error) {
	out := map[int64][]db.Skill{}
	if len(ids) == 0 {
		return out, nil
	}

	links, err := m.repo.TestAssignmentsSkillsByFilters(ctx,
		&db.TestAssignmentsSkillSearch{TestAssignmentIDs: ids},
		db.PagerNoLimit,
		db.WithRelations(db.Columns.TestAssignmentsSkill.Skill),
		db.WithSort(db.SortField{Column: db.Columns.TestAssignmentsSkill.SkillID, Direction: db.SortAsc}),
	)
	if err != nil {
		return nil, err
	}

	for _, l := range links {
		if l.Skill == nil {
			m.Error(ctx, "testAssignmentsSkill without a skill", "testAssignmentId", l.TestAssignmentID, "skillId", l.SkillID)

			continue
		}

		out[l.TestAssignmentID] = append(out[l.TestAssignmentID], *l.Skill)
	}

	return out, nil
}
