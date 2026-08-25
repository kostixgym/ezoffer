package rpc

const (
	defaultPageSize = 25
	maxPageSize     = 100
)

func normalizePager(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}

	switch {
	case pageSize <= 0:
		pageSize = defaultPageSize
	case pageSize > maxPageSize:
		pageSize = maxPageSize
	}

	return page, pageSize
}
