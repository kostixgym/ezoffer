package rpc

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"ezoffer/pkg/db"
	"ezoffer/pkg/ezoffer"

	"github.com/vmkteam/embedlog"
	"github.com/vmkteam/zenrpc/v2"
)

// listError turns a domain error into what the client should see. A bad filter
// value is the caller's mistake and is worth naming; everything else is ours and
// stays opaque, with the detail going to the log instead.
func listError(ctx context.Context, log embedlog.Logger, msg string, err error) error {
	var ve ezoffer.ValidationError
	if errors.As(err, &ve) {
		return zenrpc.NewStringError(http.StatusBadRequest, ve.Error())
	}

	log.Error(ctx, msg, "err", err)

	return ErrInternal
}

// excerptRunes is how much of the statement a list carries. Enough to tell two
// entries apart, small enough to keep a page of 100 in the tens of kilobytes.
const excerptRunes = 200

// Company and Skill are shared by the task, test assignment and interview
// catalogs — all three show the same shallow dictionary entry.
type Company struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Skill struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewCompanies(in []db.Company) []Company {
	out := make([]Company, 0, len(in))
	for _, c := range in {
		out = append(out, Company{ID: c.ID, Name: c.Name})
	}

	return out
}

func NewSkills(in []db.Skill) []Skill {
	out := make([]Skill, 0, len(in))
	for _, s := range in {
		out = append(out, Skill{ID: s.ID, Name: s.Name})
	}

	return out
}

// stringSlice makes sure an empty array reaches the client as [] and not null.
// go-pg leaves an empty text[] as a nil slice, and a client iterating the field
// should not have to special-case that.
func stringSlice(in []string) []string {
	if in == nil {
		return []string{}
	}

	return in
}

// excerpt shortens a markdown statement for a list.
//
// Most statements carry a fenced code block, so cutting by length alone would
// hand the client a half open fence. Everything from the first fence on is
// dropped, and what is left is trimmed on a word boundary.
func excerpt(content string) string {
	if head := shorten(beforeFence(content)); head != "" {
		return head
	}

	// A handful of statements open straight with a code block, leaving no prose
	// to show. Better a preview of the code with its fences stripped than a card
	// with an empty body.
	return shorten(stripFences(content))
}

func beforeFence(content string) string {
	if i := strings.Index(content, "```"); i >= 0 {
		return content[:i]
	}

	return content
}

func stripFences(content string) string {
	lines := strings.Split(content, "\n")
	kept := lines[:0]

	for _, l := range lines {
		if !strings.HasPrefix(strings.TrimSpace(l), "```") {
			kept = append(kept, l)
		}
	}

	return strings.Join(kept, "\n")
}

func shorten(content string) string {
	content = strings.TrimSpace(content)

	runes := []rune(content)
	if len(runes) <= excerptRunes {
		return content
	}

	cut := string(runes[:excerptRunes])
	if i := strings.LastIndexAny(cut, " \n\t"); i > 0 {
		cut = cut[:i]
	}

	return strings.TrimSpace(cut) + "…"
}
