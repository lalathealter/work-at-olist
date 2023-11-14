package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lalathealter/olist/db"
)

const pageQuery = "page"
const likeQuery = "like"
const GET_AUTHORS_PAGE_LIMIT = 50

func HandleGetAuthors(c *gin.Context) {
	page := c.Query(pageQuery)
	pageNum := 1
	if page != "" {
		reqPage, err := strconv.Atoi(page)
		if err == nil {
			pageNum = reqPage
		}
	}
	searchName := c.Query(likeQuery)

	dbi := db.Use()
	var authors []db.Author

	dbi.Limit(GET_AUTHORS_PAGE_LIMIT).
		Offset(GET_AUTHORS_PAGE_LIMIT*(pageNum-1)).
		Where("name LIKE ?", searchName+"%").
		Find(&authors)

	nextPageNum := pageNum + 1
	if len(authors) <= 0 {
		nextPageNum = -1
	}

	c.JSON(http.StatusOK, gin.H{
		"authors":  authors,
		"page":     pageNum,
		"nextPage": nextPageNum,
	})
}

var ErrPathParameterIsntInteger = errors.New("Path parameter should be a positive integer")

func HandleGetSingleAuthor(c *gin.Context) {
	idparam := c.Param("id")
	id64, err := strconv.ParseUint(idparam, 10, 64)
	id := uint(id64)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, ErrPathParameterIsntInteger)
		return
	}

	dbi := db.Use()
	authorObj := db.Author{ID: id}
	dbi.First(&authorObj)

	c.JSON(http.StatusOK, authorObj)
}
