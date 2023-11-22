package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lalathealter/olist/db"
	"gorm.io/gorm/clause"
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

const MAX_AUTHORS_FOR_BOOK = 32

var ErrTooManyAuthors = errors.New(fmt.Sprintf("A book can't have more than %d authors", MAX_AUTHORS_FOR_BOOK))

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

	if len(bookInput.Authors) > MAX_AUTHORS_FOR_BOOK {
		c.AbortWithError(http.StatusBadRequest, ErrTooManyAuthors)
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

var ErrToDeleteMustProvideBookID = errors.New("In order to delete a book you need to specify its ID")

func HandleDeleteBooks(c *gin.Context) {

	idString, isThere := c.Params.Get("id")
	if !isThere {
		c.AbortWithError(http.StatusBadRequest, ErrToDeleteMustProvideBookID)
		return
	}

	idToDel64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	idToDel := uint(idToDel64)

	dbi := db.Use()

	balToDelete := db.BookAuthorLink{BookID: idToDel}
	if err := dbi.Where(&balToDelete).Delete(&balToDelete).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	bookToDelete := db.Book{ID: idToDel}
	if dbi = dbi.Delete(&bookToDelete); dbi.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if dbi.RowsAffected < 1 {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusNoContent)
}

var ErrToUpdateMustProvideBookID = errors.New("To update a book you need to specify its id")

func HandleUpdateBooks(c *gin.Context) {
	idString, isThere := c.Params.Get("id")
	if !isThere {
		c.AbortWithError(http.StatusBadRequest, ErrToUpdateMustProvideBookID)
		return
	}

	bookId64, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	bookId := uint(bookId64)

	bookInput := BookInput{}
	if err := c.BindJSON(&bookInput); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	bookUpdates := db.Book{
		Name:    bookInput.Name,
		PubYear: bookInput.PubYear,
		Edition: bookInput.Edition,
	}

	dbi := db.Use()
	bookToUpdate := db.Book{ID: bookId}
	dbr := dbi.Where(&bookToUpdate).Updates(&bookUpdates)
	if err := dbr.Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	bal := db.BookAuthorLink{BookID: bookId}

	dbi = dbi.Where(&bal).Not("author_id IN ?", bookInput.Authors)
	dbr = dbi.Delete(&bal)
	if err := dbr.Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	balsToInsert := make([]db.BookAuthorLink, 0, MAX_AUTHORS_FOR_BOOK)
	for _, authorID := range bookInput.Authors {
		bal.AuthorID = authorID
		balsToInsert = append(balsToInsert, bal)
	}

	dbr = dbi.Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(balsToInsert, MAX_AUTHORS_FOR_BOOK)

	if err := dbr.Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}
