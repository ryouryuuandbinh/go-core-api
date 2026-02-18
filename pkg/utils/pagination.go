package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Pagination struct chứa các tham số phân trang
type Pagination struct {
	Limit   int    `json:"limit"`
	Page    int    `json:"page"`
	Sort    string `json:"sort"`
	Keyword string `json:"keyword"`
}

// GeneratePaginationFromRequest lấy tham số từ URLQuery

func GeneratePaginationFromRequest(c *gin.Context) Pagination {
	// Mặc định: Limit = 10, page 1, sort = "created_at desc"
	limit := 10
	page := 1
	sort := "created_at desc"
	keyword := ""

	query := c.Request.URL.Query()

	for key, value := range query {
		queryValue := value[len(value)-1]

		switch key {
		case "limit":
			limit, _ = strconv.Atoi(queryValue)
		case "page":
			page, _ = strconv.Atoi(queryValue)
		case "sort":
			sort = queryValue
		case "keyword":
			keyword = queryValue
		}
	}

	return Pagination{
		Limit:   limit,
		Page:    page,
		Sort:    sort,
		Keyword: keyword,
	}
}

// GetOffSet tính toán vị trí bắt đầu lấy dữ liệu trong DB

func (p *Pagination) GetOffSet() int {
	return (p.Page - 1) * p.Limit
}
