package rpc

import (
	"context"
	"time"

	"ezoffer/pkg/db"
	"ezoffer/pkg/ezoffer"

	"github.com/vmkteam/embedlog"
	"github.com/vmkteam/zenrpc/v2"
)

// Video kinds. A recording is stored in one of three ways and the client needs
// to know which before it decides between an iframe and a plain link.
const (
	VideoKindYoutube  = "youtube"
	VideoKindVfs      = "vfs"
	VideoKindExternal = "external"
)

// Video flattens the three nullable location columns into one object, so the
// client never has to work out which of them is set.
type Video struct {
	Kind       string `json:"kind"`
	Url        string `json:"url"`
	Embeddable bool   `json:"embeddable"`
}

type Interview struct {
	ID            int64      `json:"id"`
	Slug          string     `json:"slug"`
	Title         string     `json:"title"`
	Grades        []string   `json:"grades"`
	Types         []string   `json:"types"`
	Company       *Company   `json:"company"`
	Video         *Video     `json:"video"`
	IsReal        bool       `json:"isReal"`
	PublishedDate *time.Time `json:"publishedDate"`
}

type InterviewList struct {
	Items      []Interview `json:"items"`
	TotalCount int         `json:"totalCount"`
}

// NewVideo picks the location the recording actually lives at. Youtube wins over
// a stored file, and a bare source link is the fallback the private t.me records
// land on. Returns nil only if all three are empty, which the table's CHECK
// forbids.
func NewVideo(in db.Interview) *Video {
	switch {
	case in.YoutubeUrl != nil && *in.YoutubeUrl != "":
		return &Video{Kind: VideoKindYoutube, Url: *in.YoutubeUrl, Embeddable: in.IsEmbeddable}
	case in.VfsPath != nil && *in.VfsPath != "":
		return &Video{Kind: VideoKindVfs, Url: *in.VfsPath, Embeddable: in.IsEmbeddable}
	case in.SourceUrl != nil && *in.SourceUrl != "":
		return &Video{Kind: VideoKindExternal, Url: *in.SourceUrl, Embeddable: false}
	}

	return nil
}

func NewInterview(in db.Interview) Interview {
	out := Interview{
		ID:            in.ID,
		Slug:          in.Slug,
		Title:         in.Title,
		Grades:        stringSlice(in.Grades),
		Types:         stringSlice(in.Types),
		Video:         NewVideo(in),
		IsReal:        in.IsReal,
		PublishedDate: in.PublishedDate,
	}

	if in.Company != nil {
		out.Company = &Company{ID: in.Company.ID, Name: in.Company.Name}
	}

	return out
}

type InterviewService struct {
	zenrpc.Service
	embedlog.Logger

	m *ezoffer.Manager
}

func NewInterviewService(m *ezoffer.Manager, logger embedlog.Logger) *InterviewService {
	return &InterviewService{Logger: logger, m: m}
}

// List returns a page of interview recordings, newest published first.
//
//zenrpc:page=1 page number, starts at 1
//zenrpc:pageSize=25 items per page, capped at 100
//zenrpc:search substring to search in title
//zenrpc:grades filter by grades: junior, middle, senior, lead
//zenrpc:types filter by types: technical, liveCoding, algorithmic, hrScreening, final, systemDesign
//zenrpc:companyID filter by company
//zenrpc:isReal keep only recordings of real interviews
//zenrpc:return interviews page with total count
//zenrpc:400 unknown grade or type
//zenrpc:500 internal error
func (s InterviewService) List(ctx context.Context, page, pageSize int, search *string, grades, types []string, companyID *int64, isReal *bool) (*InterviewList, error) {
	page, pageSize = normalizePager(page, pageSize)

	interviews, totalCount, err := s.m.Interviews(ctx, ezoffer.InterviewListParams{
		Search:    search,
		Grades:    grades,
		Types:     types,
		CompanyID: companyID,
		IsReal:    isReal,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, listError(ctx, s.Logger, "failed to list interviews", err)
	}

	items := make([]Interview, 0, len(interviews))
	for _, i := range interviews {
		items = append(items, NewInterview(i))
	}

	return &InterviewList{Items: items, TotalCount: totalCount}, nil
}

// Get returns a single interview by id.
//
//zenrpc:id interview id
//zenrpc:return interview
//zenrpc:404 interview not found
//zenrpc:500 internal error
func (s InterviewService) Get(ctx context.Context, id int64) (*Interview, error) {
	interview, err := s.m.Interview(ctx, id)
	if err != nil {
		s.Error(ctx, "failed to get interview", "err", err, "id", id)
		return nil, ErrInternal
	}

	if interview == nil {
		return nil, ErrNotFound
	}

	out := NewInterview(*interview)

	return &out, nil
}
