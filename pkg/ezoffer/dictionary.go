package ezoffer

import (
	"context"

	"ezoffer/pkg/db"
)

// Companies returns the whole company dictionary. 81 rows today, and the list is
// what the client needs to build a company filter, so it is not paged.
func (m *Manager) Companies(ctx context.Context) ([]db.Company, error) {
	return m.repo.CompaniesByFilters(ctx, &db.CompanySearch{}, db.PagerNoLimit,
		db.WithSort(db.SortField{Column: db.Columns.Company.Name, Direction: db.SortAsc}),
	)
}

// Skills returns the whole skill dictionary: 13 rows, same reasoning as Companies.
func (m *Manager) Skills(ctx context.Context) ([]db.Skill, error) {
	return m.repo.SkillsByFilters(ctx, &db.SkillSearch{}, db.PagerNoLimit,
		db.WithSort(db.SortField{Column: db.Columns.Skill.Name, Direction: db.SortAsc}),
	)
}
