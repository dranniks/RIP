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

	// Services.
	r.GET("/services", h.GetServices)
	r.GET("/services/:slug", h.GetService)

	// Claims (canonical routes used by templates).
	r.GET("/artifact_claims", h.GetClaim)
	r.GET("/artifact_claims/:code", h.GetClaim)
	r.POST("/artifact_claims/add-service", h.AddServiceToDraft)
	r.POST("/artifact_claims/delete", h.DeleteDraftClaim)
	r.POST("/artifact_claims/:code/delete", h.DeleteDraftClaim)

	// Legacy aliases kept for backward compatibility.
	r.GET("/claims", h.GetClaim)
	r.GET("/claims/:code", h.GetClaim)
	r.POST("/claims/add-service", h.AddServiceToDraft)
	r.POST("/claims/delete", h.DeleteDraftClaim)
	r.POST("/claims/:code/delete", h.DeleteDraftClaim)

	if runErr := r.Run(":8080"); runErr != nil {
		log.Printf("server stopped with error: %v", runErr)
	}

	log.Println("server down")
}
