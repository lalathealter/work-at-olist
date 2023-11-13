package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lalathealter/olist/server/controllers"
)

func main() {
	fmt.Println("Hello world!")

	server := gin.Default()
	server.GET("/authors", controllers.HandleGetAuthors)
	// server.GET("/books", HandleGetBooks)
	server.POST("/books", controllers.HandlePostBooks)
	server.Run("localhost:5050")

}

