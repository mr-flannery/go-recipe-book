package handlers

type PaginationData struct {
	CurrentPage int
	TotalPages  int
	PageSize    int
	PageNumbers []int
	TotalCount  int
	RangeStart  int
	RangeEnd    int
	HasMore     bool
	HasPrevious bool
	NextOffset  int
	PrevOffset  int
}

func CalculatePagination(totalCount, currentPage, pageSize, loadedCount int) PaginationData {
	if pageSize <= 0 {
		pageSize = 20
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	if currentPage < 1 {
		currentPage = 1
	}
	if currentPage > totalPages {
		currentPage = totalPages
	}

	pageNumbers := calculatePageNumbers(currentPage, totalPages)

	rangeStart := (currentPage-1)*pageSize + 1
	rangeEnd := rangeStart + loadedCount - 1
	if loadedCount == 0 {
		rangeStart = 0
		rangeEnd = 0
	}

	offset := (currentPage - 1) * pageSize
	hasMore := rangeEnd < totalCount
	hasPrevious := currentPage > 1

	return PaginationData{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		PageSize:    pageSize,
		PageNumbers: pageNumbers,
		TotalCount:  totalCount,
		RangeStart:  rangeStart,
		RangeEnd:    rangeEnd,
		HasMore:     hasMore,
		HasPrevious: hasPrevious,
		NextOffset:  offset + loadedCount,
		PrevOffset:  max(0, offset-pageSize),
	}
}

func calculatePageNumbers(currentPage, totalPages int) []int {
	if totalPages <= 5 {
		pages := make([]int, totalPages)
		for i := 0; i < totalPages; i++ {
			pages[i] = i + 1
		}
		return pages
	}

	start := currentPage - 2
	end := currentPage + 2

	if start < 1 {
		end += (1 - start)
		start = 1
	}
	if end > totalPages {
		start -= (end - totalPages)
		end = totalPages
	}
	if start < 1 {
		start = 1
	}

	pages := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}
