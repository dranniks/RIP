package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	log.Println("Server start up")

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/styles", "./resources/styles")

	// Optional convenience: open app root at services list.
	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/services")
	})

	// 1) List of services + search query parameter (?q=...)
	r.GET("/services", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"q": ctx.Query("q"),
		})
	})

	// 2) Service details by service ID
	r.GET("/services/:id", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "service.html", gin.H{
			"serviceID": ctx.Param("id"),
		})
	})

	// 3) Claim page by claim ID
	r.GET("/claims/:id", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "claim.html", gin.H{
			"claimID": ctx.Param("id"),
		})
	})

	if err := r.Run(); err != nil {
		log.Printf("server stopped with error: %v", err)
	}

	log.Println("Server down")
}
