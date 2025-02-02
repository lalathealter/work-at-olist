package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lalathealter/olist/server/controllers"
)

func main() {
	fmt.Println("Hello world!")

	server := gin.Default()
	server.Use(gin.ErrorLogger())
	server.GET("/authors", controllers.HandleGetAuthors)
	server.GET("/authors/:id", controllers.HandleGetSingleAuthor)
	server.GET("/books", controllers.HandleGetBooks)
	server.DELETE("/books/:id", controllers.HandleDeleteBooks)
	server.PATCH("/books/:id", controllers.HandleUpdateBooks)
	server.POST("/books", controllers.HandlePostBooks)
	server.Run("localhost:5050")

}
