package ezoffer

import (
	"context"

	"ezoffer/pkg/db"
)

func (m *Manager) Companies(ctx context.Context) ([]db.Company, error) {
	return m.repo.CompaniesByFilters(ctx, &db.CompanySearch{}, db.PagerNoLimit,
		db.WithSort(db.SortField{Column: db.Columns.Company.Name, Direction: db.SortAsc}),
	)
}

func (m *Manager) Skills(ctx context.Context) ([]db.Skill, error) {
	return m.repo.SkillsByFilters(ctx, &db.SkillSearch{}, db.PagerNoLimit,
		db.WithSort(db.SortField{Column: db.Columns.Skill.Name, Direction: db.SortAsc}),
	)
}
