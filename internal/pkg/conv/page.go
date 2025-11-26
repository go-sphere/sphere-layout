package conv

import "golang.org/x/exp/constraints"

// DefaultPageSize defines the standard page size used for pagination operations.
const DefaultPageSize = 20

// Page calculates pagination values based on total items and page size.
// It returns the number of pages needed and the effective page size to use.
// If pageSize is invalid, it uses the defaultSize parameter.
func Page[P constraints.Integer](total, pageSize, defaultSize P) (page P, size P) {
	if pageSize <= 0 {
		pageSize = max(1, defaultSize)
	}
	if total == 0 {
		return 0, pageSize
	}
	page = total / pageSize
	if total%pageSize != 0 {
		page++
	}
	return page, pageSize
}
