package handlers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type PaginationData struct {
	CurrentPage int
	TotalPages  int
	PageSize    int
	PageNumbers []int
	TotalCount  int
	RangeStart  int // First item number (1-indexed)
	RangeEnd    int // Last item number
}

func CalculatePagination(totalCount, currentPage, pageSize int) PaginationData {
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
	rangeEnd := currentPage * pageSize
	if rangeEnd > totalCount {
		rangeEnd = totalCount
	}
	if totalCount == 0 {
		rangeStart = 0
		rangeEnd = 0
	}

	return PaginationData{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		PageSize:    pageSize,
		PageNumbers: pageNumbers,
		TotalCount:  totalCount,
		RangeStart:  rangeStart,
		RangeEnd:    rangeEnd,
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

type FilterState struct {
	Page          int
	PageSize      int
	Search        string
	Tags          string
	UserTags      string
	AuthoredByMe  bool
	CaloriesOp    string
	CaloriesValue int
	PrepTimeOp    string
	PrepTimeValue int
	CookTimeOp    string
	CookTimeValue int
}

func (f FilterState) ToURLQuery() string {
	params := url.Values{}

	if f.Page > 1 {
		params.Set("page", strconv.Itoa(f.Page))
	}
	if f.PageSize > 0 && f.PageSize != 20 {
		params.Set("page_size", strconv.Itoa(f.PageSize))
	}
	if f.Search != "" {
		params.Set("search", f.Search)
	}
	if f.Tags != "" {
		params.Set("tags", f.Tags)
	}
	if f.UserTags != "" {
		params.Set("user_tags", f.UserTags)
	}
	if f.AuthoredByMe {
		params.Set("authored_by_me", "1")
	}
	if f.CaloriesValue > 0 && f.CaloriesOp != "" {
		params.Set("calories_op", f.CaloriesOp)
		params.Set("calories_value", strconv.Itoa(f.CaloriesValue))
	}
	if f.PrepTimeValue > 0 && f.PrepTimeOp != "" {
		params.Set("prep_time_op", f.PrepTimeOp)
		params.Set("prep_time_value", strconv.Itoa(f.PrepTimeValue))
	}
	if f.CookTimeValue > 0 && f.CookTimeOp != "" {
		params.Set("cook_time_op", f.CookTimeOp)
		params.Set("cook_time_value", strconv.Itoa(f.CookTimeValue))
	}

	query := params.Encode()
	if query == "" {
		return "/recipes"
	}
	return "/recipes?" + query
}

func ParseFilterStateFromQuery(values url.Values) FilterState {
	state := FilterState{
		Page:     1,
		PageSize: 20,
	}

	if p, err := strconv.Atoi(values.Get("page")); err == nil && p > 0 {
		state.Page = p
	}
	if ps, err := strconv.Atoi(values.Get("page_size")); err == nil && ps > 0 {
		state.PageSize = ps
	}
	state.Search = strings.TrimSpace(values.Get("search"))
	state.Tags = values.Get("tags")
	state.UserTags = values.Get("user_tags")
	state.AuthoredByMe = values.Get("authored_by_me") == "1"

	if cv, err := strconv.Atoi(values.Get("calories_value")); err == nil && cv > 0 {
		state.CaloriesValue = cv
		state.CaloriesOp = values.Get("calories_op")
	}
	if pv, err := strconv.Atoi(values.Get("prep_time_value")); err == nil && pv > 0 {
		state.PrepTimeValue = pv
		state.PrepTimeOp = values.Get("prep_time_op")
	}
	if cv, err := strconv.Atoi(values.Get("cook_time_value")); err == nil && cv > 0 {
		state.CookTimeValue = cv
		state.CookTimeOp = values.Get("cook_time_op")
	}

	return state
}

func formatPagesParam(startPage, endPage int) string {
	return fmt.Sprintf("%d-%d", startPage, endPage)
}
