package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lalathealter/olist/db"
)

type BookInput struct {
	Name    string `binding:"required"`
	Edition uint   `binding:"required"`
	PubYear int    `json:"publication_year" binding:"required"`
	Authors []uint `binding:"required"`
}

func HandlePostBooks(c *gin.Context) {

	bookInput := BookInput{}
	if err := c.BindJSON(&bookInput); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	dbi := db.Use()
	newBook := db.Book{
		Name:    bookInput.Name,
		PubYear: bookInput.PubYear,
		Edition: bookInput.Edition,
	}
	dbi.Create(&newBook)

	for _, author := range bookInput.Authors {
		balink := db.BookAuthorLink{
			BookID:   newBook.ID,
			AuthorID: author,
		}
		dbi.Create(&balink)
	}

	c.Status(http.StatusCreated)
}
