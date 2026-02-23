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
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	return Pagination{
		Limit:   limit,
		Page:    page,
		Sort:    c.DefaultQuery("sort", "created_at desc"),
		Keyword: c.Query("keyword"),
	}
}

// GetOffSet tính toán vị trí bắt đầu lấy dữ liệu trong DB

func (p *Pagination) GetOffSet() int {
	return (p.Page - 1) * p.Limit
}
