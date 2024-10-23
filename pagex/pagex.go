package pagex

type Pagination struct {
	Page     int64
	PageSize int64
	MaxCount int64
}

// Generate new pagination
func (p *Pagination) NewPagination(page, pageSize, count int64) (pagination Pagination) {
	if page == 0 {
		pagination.Page = 1
	} else {
		pagination.Page = page
	}

	if pageSize == 0 {
		pagination.PageSize = 15
	} else {
		pagination.PageSize = pageSize
	}

	pagination.MaxCount = count
	return
}

// Offset Returns the current offset.
func (p *Pagination) Offset() int64 {
	return (p.Page - 1) * p.PageSize
}
