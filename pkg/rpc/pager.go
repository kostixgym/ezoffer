package rpc

const (
	defaultPageSize = 25
	maxPageSize     = 100

	// Offset is computed as (page-1)*pageSize. Without an upper bound that
	// multiplication overflows int64 and wraps into a negative OFFSET, which
	// Postgres rejects — a 500 anyone can trigger with a single request.
	maxPage = 100000
)

func normalizePager(page, pageSize int) (int, int) {
	switch {
	case page < 1:
		page = 1
	case page > maxPage:
		page = maxPage
	}

	switch {
	case pageSize <= 0:
		pageSize = defaultPageSize
	case pageSize > maxPageSize:
		pageSize = maxPageSize
	}

	return page, pageSize
}
