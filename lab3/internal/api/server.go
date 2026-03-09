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

	api := r.Group("/api")
	{
		api.GET("/services", h.GetServicesAPI)
		api.GET("/services/:id", h.GetServiceAPI)
		api.POST("/services", h.CreateServiceAPI)

		api.POST("/claim-items", h.AddServiceToDraftAPI)
		api.PUT("/claim-items/:service_id", h.UpdateDraftMatchAPI)
		api.DELETE("/claim-items/:service_id", h.DeleteDraftMatchAPI)

		api.GET("/claims/cart-icon", h.GetCartIconAPI)
		api.GET("/claims", h.GetClaimsAPI)
		api.GET("/claims/:id", h.GetClaimAPI)
		api.PUT("/claims/:id", h.UpdateDraftClaimAPI)
		api.PUT("/claims/:id/form", h.FormClaimAPI)
		api.PUT("/claims/:id/moderate", h.ModerateClaimAPI)
		api.DELETE("/claims/:id", h.DeleteDraftClaimAPI)

		api.POST("/users/register", h.RegisterUserAPI)
		api.POST("/users/auth", h.AuthStubAPI)
		api.POST("/users/logout", h.LogoutStubAPI)
	}

	if runErr := r.Run(":8080"); runErr != nil {
		log.Printf("server stopped with error: %v", runErr)
	}

	log.Println("server down")
}
