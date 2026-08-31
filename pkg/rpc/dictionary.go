package rpc

import (
	"context"

	"ezoffer/pkg/ezoffer"

	"github.com/vmkteam/embedlog"
	"github.com/vmkteam/zenrpc/v2"
)

// DictionaryService hands out the lists a client needs before it can build a
// filter: the catalogs take companyID and skillIDs, and nothing else tells the
// client which ids exist.
type DictionaryService struct {
	zenrpc.Service
	embedlog.Logger

	m *ezoffer.Manager
}

func NewDictionaryService(m *ezoffer.Manager, logger embedlog.Logger) *DictionaryService {
	return &DictionaryService{Logger: logger, m: m}
}

// Companies returns every company, ordered by name.
//
//zenrpc:return companies
//zenrpc:500 internal error
func (s DictionaryService) Companies(ctx context.Context) ([]Company, error) {
	companies, err := s.m.Companies(ctx)
	if err != nil {
		s.Error(ctx, "failed to list companies", "err", err)
		return nil, ErrInternal
	}

	return NewCompanies(companies), nil
}

// Skills returns every skill, ordered by name.
//
//zenrpc:return skills
//zenrpc:500 internal error
func (s DictionaryService) Skills(ctx context.Context) ([]Skill, error) {
	skills, err := s.m.Skills(ctx)
	if err != nil {
		s.Error(ctx, "failed to list skills", "err", err)
		return nil, ErrInternal
	}

	return NewSkills(skills), nil
}
