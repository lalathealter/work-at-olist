package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lalathealter/olist/db"
)

const likeQuery = "like"

func HandleGetAuthors(c *gin.Context) {
	pagVals := ParsePagination(c)
	searchName := c.Query(likeQuery)

	dbi := db.Use()
	var authors []db.Author

	dbi.Limit(int(pagVals.Limit)).
		Offset(int(pagVals.Limit)*int(pagVals.Page-1)).
		Where("name LIKE ?", searchName+"%").
		Find(&authors)

	if len(authors) <= 0 {
		pagVals.NextPage = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"authors":    authors,
		"pagination": pagVals,
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
