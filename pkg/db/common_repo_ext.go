package db

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
)

// IncCommentLikes bumps the like counter by one and returns the new value.
// Returns nil when there is no visible comment with that id.
//
// The increment happens inside the statement on purpose. Reading the counter and
// writing back value+1 through UpdateComment would drop every like that lands
// between the read and the write.
func (cr CommonRepo) IncCommentLikes(ctx context.Context, id int64) (*int64, error) {
	var likesCnt int64

	_, err := cr.db.QueryOneContext(ctx, pg.Scan(&likesCnt),
		`UPDATE "comments" SET "likesCnt" = "likesCnt" + 1 WHERE "commentId" = ? AND "isVisible" RETURNING "likesCnt"`,
		id,
	)
	if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &likesCnt, nil
}
