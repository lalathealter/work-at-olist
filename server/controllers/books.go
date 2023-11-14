package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lalathealter/olist/db"
)

const (
	nameQuery    = "name"
	pubYearQuery = "publication_year"
	editionQuery = "edition"
	authorQuery  = "author"
)

func HandleGetBooks(c *gin.Context) {
	name := c.Query(nameQuery)
	pubYearStr := c.Query(pubYearQuery)
	editionStr := c.Query(editionQuery)
	authorStr := c.Query(authorQuery)

	pubYear, errPubYearConv := strconv.Atoi(pubYearStr)

	editionUint64, errEditionConv := strconv.ParseUint(editionStr, 10, 64)
	var edition uint
	if errEditionConv == nil {
		edition = uint(editionUint64)
	}

	authorUint64, errAuthorConv := strconv.ParseUint(authorStr, 10, 64)
	var authorId uint
	if errAuthorConv == nil {
		authorId = uint(authorUint64)
	}

	dbi := db.Use()

	bookDetails := db.Book{}
	if name != "" {
		bookDetails.Name = name
	}
	if errPubYearConv == nil {
		bookDetails.PubYear = pubYear
	}
	if errEditionConv == nil {
		bookDetails.Edition = edition
	}

	if errAuthorConv == nil {
		dbi = dbi.Model(&db.Book{}).Where(
			"(ID) IN (?)",
			dbi.
				Model(&db.BookAuthorLink{}).
				Select("BookID").
				Where("Author_ID = ?", authorId),
		)
	}

	out := make([]*db.BookWithAuthors, 0, 5)
	authorsSub := db.Use().Model(&db.BookAuthorLink{}).
		Select("book_id", "ARRAY_AGG(author_id) AS authors").
		Group("book_id")

	dbi.Select("authors", "name", "edition", "pub_year", "ID").
		Joins("INNER JOIN (?) auths ON books.ID = auths.book_id", authorsSub).
		Find(&out, &bookDetails)

	c.JSON(http.StatusOK, out)
}

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

	if dbc := dbi.Create(&newBook); dbc.Error != nil {
		c.AbortWithError(http.StatusBadRequest, dbc.Error)
		return
	}

	for _, author := range bookInput.Authors {
		balink := db.BookAuthorLink{
			BookID:   newBook.ID,
			AuthorID: author,
		}

		if dbc := dbi.Create(&balink); dbc.Error != nil {
			c.AbortWithError(http.StatusInternalServerError, dbc.Error)
			return
		}
	}

	c.Status(http.StatusCreated)
}
