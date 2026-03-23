package api

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"xrfApp/internal/app/auth"
	"xrfApp/internal/app/handler"
	"xrfApp/internal/app/middleware"
	"xrfApp/internal/app/repository"
)

func StartServer() {
	log.Println("server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		log.Fatalf("cannot init repository: %v", err)
	}
	tokenManager := auth.NewManagerFromEnv()
	h := handler.NewHandler(repo, tokenManager)

	r := gin.Default()

	r.GET("/swagger", swaggerUIHandler)
	r.GET("/swagger/", swaggerUIHandler)
	r.GET("/swagger/openapi.json", openAPISpecHandler)

	api := r.Group("/api")
	{
		public := api.Group("")
		{
			public.GET("/services", h.GetServicesAPI)
			public.GET("/services/:id", h.GetServiceAPI)
			public.POST("/users/register", h.RegisterUserAPI)
			public.POST("/users/auth", h.AuthAPI)
		}

		authenticated := api.Group("")
		authenticated.Use(middleware.RequireAuth(tokenManager))
		{
			authenticated.POST("/services", h.CreateServiceAPI)

			authenticated.POST("/claim-items", h.AddServiceToDraftAPI)
			authenticated.PUT("/claim-items/:service_id", h.UpdateDraftMatchAPI)
			authenticated.DELETE("/claim-items/:service_id", h.DeleteDraftMatchAPI)

			authenticated.GET("/claims/cart-icon", h.GetCartIconAPI)
			authenticated.GET("/claims", h.GetClaimsAPI)
			authenticated.GET("/claims/:id", h.GetClaimAPI)
			authenticated.PUT("/claims/:id", h.UpdateDraftClaimAPI)
			authenticated.PUT("/claims/:id/form", h.FormClaimAPI)
			authenticated.DELETE("/claims/:id", h.DeleteDraftClaimAPI)

			authenticated.POST("/users/logout", h.LogoutStubAPI)
		}

		moderator := api.Group("")
		moderator.Use(middleware.RequireAuth(tokenManager), middleware.RequireRoles("moderator"))
		{
			moderator.PUT("/claims/:id/moderate", h.ModerateClaimAPI)
		}
	}

	port := strings.TrimSpace(os.Getenv("APP_PORT"))
	if port == "" {
		port = "8080"
	}

	if runErr := r.Run(":" + port); runErr != nil {
		log.Printf("server stopped with error: %v", runErr)
	}

	log.Println("server down")
}
