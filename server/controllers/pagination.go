package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const pageQuery = "page"
const limitQuery = "limit"

type PaginationValues struct {
	Limit    uint64
	Page     uint64
	NextPage uint64
}

var DefaultPaginationValues = PaginationValues{50, 1, 2}

func ParsePagination(c *gin.Context) PaginationValues {

	ps := DefaultPaginationValues

	if pageStr := c.Query(pageQuery); pageStr != "" {
		uintPage, err := strconv.ParseUint(pageStr, 10, 64)
		if err == nil && uintPage >= 1 {
			ps.Page = uintPage
			ps.NextPage = uintPage + 1
			if ps.NextPage < ps.Page {
				ps.NextPage = 0
			}
		}
	}

	if limitStr := c.Query(limitQuery); limitStr != "" {
		limitUint, err := strconv.ParseUint(limitStr, 10, 64)
		if err == nil && limitUint >= 1 {
			ps.Limit = limitUint
		}
	}
	return ps
}
