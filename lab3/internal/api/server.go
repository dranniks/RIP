package api

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"xrfApp/internal/app/auth"
	"xrfApp/internal/app/handler"
	"xrfApp/internal/app/middleware"
	"xrfApp/internal/app/repository"
	"xrfApp/internal/app/session"
)

func StartServer() {
	log.Println("server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		log.Fatalf("cannot init repository: %v", err)
	}
	tokenManager := auth.NewManagerFromEnv()
	sessionManager := session.NewManagerFromEnv()
	if err := sessionManager.Ping(context.Background()); err != nil {
		log.Fatalf("cannot init redis token store: %v", err)
	}

	h := handler.NewHandler(repo, tokenManager, sessionManager)

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
			public.GET("/claims/cart-icon", h.GetCartIconAPI)
			public.POST("/users/register", h.RegisterUserAPI)
			public.POST("/users/auth", h.AuthAPI)
		}

		authenticated := api.Group("")
		authenticated.Use(middleware.RequireAuth(tokenManager, sessionManager))
		{
			authenticated.POST("/services", h.CreateServiceAPI)

			authenticated.POST("/claim-items", h.AddServiceToDraftAPI)
			authenticated.PUT("/claim-items/:service_id", h.UpdateDraftMatchAPI)
			authenticated.DELETE("/claim-items/:service_id", h.DeleteDraftMatchAPI)

			authenticated.GET("/claims", h.GetClaimsAPI)
			authenticated.GET("/claims/:id", h.GetClaimAPI)
			authenticated.PUT("/claims/:id", h.UpdateDraftClaimAPI)
			authenticated.PUT("/claims/:id/form", h.FormClaimAPI)
			authenticated.DELETE("/claims/:id", h.DeleteDraftClaimAPI)

			authenticated.POST("/users/logout", h.LogoutStubAPI)
		}

		moderator := api.Group("")
		moderator.Use(middleware.RequireAuth(tokenManager, sessionManager), middleware.RequireRoles("moderator"))
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
