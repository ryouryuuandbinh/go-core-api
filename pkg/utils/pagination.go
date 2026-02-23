package utils

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Limit   int    `json:"limit"`
	Page    int    `json:"page"`
	Sort    string `json:"sort"`
	Keyword string `json:"keyword"`
}

func GeneratePaginationFromRequest(c *gin.Context) Pagination {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	// BẢO MẬT: Chặn DoS bằng cách ép limit tối đa
	if err != nil || limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	// BẢO MẬT: Validate chống SQL Injection cho tham số Sort
	rawSort := c.DefaultQuery("sort", "created_at desc")
	safeSort := validateSortQuery(rawSort)

	return Pagination{
		Limit:   limit,
		Page:    page,
		Sort:    safeSort,
		Keyword: c.Query("keyword"),
	}
}

// validateSortQuery chỉ cho phép sort các cột an toàn
func validateSortQuery(sort string) string {
	allowedColumns := map[string]bool{
		"id": true, "created_at": true, "updated_at": true, "email": true,
	}
	parts := strings.Split(strings.TrimSpace(sort), " ")
	if len(parts) == 0 || len(parts) > 2 {
		return "created_at desc"
	}

	col := parts[0]
	if !allowedColumns[col] {
		return "created_at desc" // Mặc định nếu phát hiện cột lạ
	}

	direction := "desc"
	if len(parts) == 2 && strings.ToLower(parts[1]) == "asc" {
		direction = "asc"
	}

	return col + " " + direction
}

func (p *Pagination) GetOffSet() int {
	return (p.Page - 1) * p.Limit
}
