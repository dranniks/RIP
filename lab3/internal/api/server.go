package api

import (
	"log"

	"github.com/gin-gonic/gin"

	"xrfApp/internal/app/handler"
	"xrfApp/internal/app/repository"
)

func StartServer() {
	log.Println("server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		log.Fatalf("cannot init repository: %v", err)
	}
	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/styles", "./resources/styles")

	// 3 GET routes.
	r.GET("/services", h.GetServices)
	r.GET("/services/:slug", h.GetService)
	r.GET("/claims/:code", h.GetClaim)

	// 2 POST routes.
	r.POST("/claims/add-service", h.AddServiceToDraft)
	r.POST("/claims/:code/delete", h.DeleteDraftClaim)

	if runErr := r.Run(":8080"); runErr != nil {
		log.Printf("server stopped with error: %v", runErr)
	}

	log.Println("server down")
}
