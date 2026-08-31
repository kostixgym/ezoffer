package rpc

import (
	"context"
	"time"

	"ezoffer/pkg/ezoffer"

	"github.com/vmkteam/embedlog"
	"github.com/vmkteam/zenrpc/v2"
)

// TaskSummary is a list entry: the statement is replaced by a short excerpt.
type TaskSummary struct {
	ID        int64      `json:"id"`
	Slug      string     `json:"slug"`
	Title     string     `json:"title"`
	Grades    []string   `json:"grades"`
	Type      *string    `json:"type"`
	Companies []Company  `json:"companies"`
	Excerpt   string     `json:"excerpt"`
	LastDate  *time.Time `json:"lastDate"`
}

// Task is the full card, with the statement and the source link.
type Task struct {
	ID        int64      `json:"id"`
	Slug      string     `json:"slug"`
	Title     string     `json:"title"`
	Grades    []string   `json:"grades"`
	Type      *string    `json:"type"`
	Companies []Company  `json:"companies"`
	Content   string     `json:"content"`
	SourceUrl *string    `json:"sourceUrl"`
	LastDate  *time.Time `json:"lastDate"`
}

type TaskList struct {
	Items      []TaskSummary `json:"items"`
	TotalCount int           `json:"totalCount"`
}

func NewTaskSummary(in ezoffer.TaskItem) TaskSummary {
	return TaskSummary{
		ID:        in.Task.ID,
		Slug:      in.Task.Slug,
		Title:     in.Task.Title,
		Grades:    stringSlice(in.Task.Grades),
		Type:      in.Task.Type,
		Companies: NewCompanies(in.Companies),
		Excerpt:   excerpt(in.Task.Content),
		LastDate:  in.Task.LastDate,
	}
}

func NewTask(in ezoffer.TaskItem) Task {
	return Task{
		ID:        in.Task.ID,
		Slug:      in.Task.Slug,
		Title:     in.Task.Title,
		Grades:    stringSlice(in.Task.Grades),
		Type:      in.Task.Type,
		Companies: NewCompanies(in.Companies),
		Content:   in.Task.Content,
		SourceUrl: in.Task.SourceUrl,
		LastDate:  in.Task.LastDate,
	}
}

type TaskService struct {
	zenrpc.Service
	embedlog.Logger

	m *ezoffer.Manager
}

func NewTaskService(m *ezoffer.Manager, logger embedlog.Logger) *TaskService {
	return &TaskService{Logger: logger, m: m}
}

// List returns a page of live coding tasks.
//
//zenrpc:page=1 page number, starts at 1
//zenrpc:pageSize=25 items per page, capped at 100
//zenrpc:search substring to search in task title
//zenrpc:grades filter by grades: junior, middle, senior, lead
//zenrpc:taskType filter by type: liveCoding, algorithms or systemDesign
//zenrpc:companyID filter by company
//zenrpc:return tasks page with total count
//zenrpc:400 unknown grade or type
//zenrpc:500 internal error
func (s TaskService) List(ctx context.Context, page, pageSize int, search *string, grades []string, taskType *string, companyID *int64) (*TaskList, error) {
	page, pageSize = normalizePager(page, pageSize)

	tasks, totalCount, err := s.m.Tasks(ctx, ezoffer.TaskListParams{
		Search:    search,
		Grades:    grades,
		Type:      taskType,
		CompanyID: companyID,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, listError(ctx, s.Logger, "failed to list tasks", err)
	}

	items := make([]TaskSummary, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, NewTaskSummary(t))
	}

	return &TaskList{Items: items, TotalCount: totalCount}, nil
}

// Get returns a single task by id.
//
//zenrpc:id task id
//zenrpc:return task
//zenrpc:404 task not found
//zenrpc:500 internal error
func (s TaskService) Get(ctx context.Context, id int64) (*Task, error) {
	task, err := s.m.Task(ctx, id)
	if err != nil {
		s.Error(ctx, "failed to get task", "err", err, "id", id)
		return nil, ErrInternal
	}

	if task == nil {
		return nil, ErrNotFound
	}

	out := NewTask(*task)

	return &out, nil
}
