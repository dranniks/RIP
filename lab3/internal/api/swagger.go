package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func swaggerUIHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
}

func openAPISpecHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, buildOpenAPISpec())
}

func buildOpenAPISpec() gin.H {
	securityBearer := []gin.H{{"BearerAuth": []string{}}}

	return gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "XRF Lab4 API",
			"version":     "1.0.0",
			"description": "REST API with JWT authorization and role permissions (creator/moderator).",
		},
		"servers": []gin.H{
			{"url": "http://localhost:8080"},
		},
		"tags": []gin.H{
			{"name": "Auth"},
			{"name": "Services"},
			{"name": "ClaimItems"},
			{"name": "Claims"},
			{"name": "Users"},
		},
		"components": gin.H{
			"securitySchemes": gin.H{
				"BearerAuth": gin.H{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
			"schemas": gin.H{
				"ErrorResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"message": gin.H{"type": "string"},
					},
				},
				"AuthRequest": gin.H{
					"type":     "object",
					"required": []string{"login", "password"},
					"properties": gin.H{
						"login":    gin.H{"type": "string"},
						"password": gin.H{"type": "string"},
					},
				},
				"RegisterRequest": gin.H{
					"type":     "object",
					"required": []string{"login", "full_name", "password"},
					"properties": gin.H{
						"login":     gin.H{"type": "string"},
						"full_name": gin.H{"type": "string"},
						"password":  gin.H{"type": "string"},
					},
				},
				"AuthResponseData": gin.H{
					"type": "object",
					"properties": gin.H{
						"user_id":    gin.H{"type": "integer"},
						"login":      gin.H{"type": "string"},
						"full_name":  gin.H{"type": "string"},
						"role":       gin.H{"type": "string"},
						"token_type": gin.H{"type": "string", "example": "Bearer"},
						"token":      gin.H{"type": "string"},
						"expires_at": gin.H{"type": "string", "format": "date-time"},
						"token_ttl":  gin.H{"type": "integer"},
						"token_expires_at": gin.H{
							"type":   "string",
							"format": "date-time",
						},
						"auth_method": gin.H{"type": "string", "example": "jwt"},
					},
				},
			},
		},
		"paths": gin.H{
			"/api/users/auth": gin.H{
				"post": gin.H{
					"tags":        []string{"Auth"},
					"summary":     "Authenticate and get JWT token",
					"description": "Returns JWT for Authorization: Bearer <token>.",
					"requestBody": gin.H{
						"required": true,
						"content": gin.H{
							"application/json": gin.H{
								"schema": gin.H{"$ref": "#/components/schemas/AuthRequest"},
							},
						},
					},
					"responses": gin.H{
						"200": gin.H{
							"description": "Authenticated",
							"content": gin.H{
								"application/json": gin.H{
									"schema": gin.H{
										"type": "object",
										"properties": gin.H{
											"data": gin.H{"$ref": "#/components/schemas/AuthResponseData"},
										},
									},
								},
							},
						},
						"400": gin.H{"description": "Validation error"},
					},
				},
			},
			"/api/users/register": gin.H{
				"post": gin.H{
					"tags":    []string{"Users"},
					"summary": "Register user",
					"requestBody": gin.H{
						"required": true,
						"content": gin.H{
							"application/json": gin.H{
								"schema": gin.H{"$ref": "#/components/schemas/RegisterRequest"},
							},
						},
					},
					"responses": gin.H{
						"201": gin.H{"description": "Created"},
						"400": gin.H{"description": "Validation error"},
					},
				},
			},
			"/api/users/logout": gin.H{
				"post": gin.H{
					"tags":        []string{"Auth"},
					"summary":     "Logout",
					"description": "Adds current JWT to Redis blacklist until token expiration.",
					"security":    securityBearer,
					"responses": gin.H{
						"200": gin.H{"description": "OK"},
						"401": gin.H{"description": "Unauthorized"},
					},
				},
			},
			"/api/services": gin.H{
				"get": gin.H{
					"tags":        []string{"Services"},
					"summary":     "List services",
					"description": "Public read endpoint. Supports filter: q.",
					"parameters": []gin.H{
						{
							"name":        "q",
							"in":          "query",
							"description": "Search filter",
							"schema":      gin.H{"type": "string"},
						},
					},
					"responses": gin.H{
						"200": gin.H{"description": "OK"},
					},
				},
				"post": gin.H{
					"tags":        []string{"Services"},
					"summary":     "Create service",
					"description": "Requires JWT.",
					"security":    securityBearer,
					"requestBody": gin.H{
						"required": true,
						"content": gin.H{
							"multipart/form-data": gin.H{
								"schema": gin.H{
									"type": "object",
									"properties": gin.H{
										"name":                gin.H{"type": "string"},
										"description":         gin.H{"type": "string"},
										"clip_description_en": gin.H{"type": "string", "description": "Short English CLIP text (50-100 chars)"},
										"era":                 gin.H{"type": "string"},
										"culture":             gin.H{"type": "string"},
										"unit_price":          gin.H{"type": "number"},
										"cu_reference":        gin.H{"type": "number"},
										"zn_reference":        gin.H{"type": "number"},
										"sn_reference":        gin.H{"type": "number"},
										"pb_reference":        gin.H{"type": "number"},
										"image":               gin.H{"type": "string", "format": "binary"},
										"video":               gin.H{"type": "string", "format": "binary"},
									},
								},
							},
						},
					},
					"responses": gin.H{
						"201": gin.H{"description": "Created"},
						"401": gin.H{"description": "Unauthorized"},
					},
				},
			},
			"/api/services/{id}": gin.H{
				"get": gin.H{
					"tags":    []string{"Services"},
					"summary": "Get service by id",
					"parameters": []gin.H{
						{
							"name":     "id",
							"in":       "path",
							"required": true,
							"schema":   gin.H{"type": "integer"},
						},
					},
					"responses": gin.H{"200": gin.H{"description": "OK"}},
				},
			},
			"/api/claim-items": gin.H{
				"post": gin.H{
					"tags":     []string{"ClaimItems"},
					"summary":  "Add service to draft claim",
					"security": securityBearer,
					"responses": gin.H{
						"201": gin.H{"description": "Created"},
						"401": gin.H{"description": "Unauthorized"},
					},
				},
			},
			"/api/claim-items/{service_id}": gin.H{
				"put": gin.H{
					"tags":     []string{"ClaimItems"},
					"summary":  "Update m-m row in draft",
					"security": securityBearer,
					"responses": gin.H{
						"200": gin.H{"description": "OK"},
						"401": gin.H{"description": "Unauthorized"},
					},
				},
				"delete": gin.H{
					"tags":     []string{"ClaimItems"},
					"summary":  "Delete service from draft",
					"security": securityBearer,
					"responses": gin.H{
						"200": gin.H{"description": "OK"},
						"401": gin.H{"description": "Unauthorized"},
					},
				},
			},
			"/api/claims/cart-icon": gin.H{
				"get": gin.H{
					"tags":        []string{"Claims"},
					"summary":     "Get cart icon",
					"description": "Public endpoint. Returns 200 for guest and creator draft service count for authorized user.",
					"responses":   gin.H{"200": gin.H{"description": "OK"}},
				},
			},
			"/api/claims": gin.H{
				"get": gin.H{
					"tags":        []string{"Claims"},
					"summary":     "List claims",
					"description": "Guest -> 401. Creator sees only own claims. Moderator sees all claims.",
					"security":    securityBearer,
					"responses": gin.H{
						"200": gin.H{"description": "OK"},
						"401": gin.H{"description": "Unauthorized"},
					},
				},
			},
			"/api/claims/{id}": gin.H{
				"get": gin.H{
					"tags":        []string{"Claims"},
					"summary":     "Get claim by id",
					"description": "Creator can read only own claim.",
					"security":    securityBearer,
					"responses":   gin.H{"200": gin.H{"description": "OK"}},
				},
				"put": gin.H{
					"tags":        []string{"Claims"},
					"summary":     "Update draft claim fields",
					"description": "Creator only for own draft.",
					"security":    securityBearer,
					"responses":   gin.H{"200": gin.H{"description": "OK"}},
				},
				"delete": gin.H{
					"tags":        []string{"Claims"},
					"summary":     "Delete draft claim",
					"description": "Creator only for own draft.",
					"security":    securityBearer,
					"responses":   gin.H{"200": gin.H{"description": "OK"}},
				},
			},
			"/api/claims/{id}/form": gin.H{
				"put": gin.H{
					"tags":        []string{"Claims"},
					"summary":     "Form draft claim",
					"description": "Creator only for own draft.",
					"security":    securityBearer,
					"responses":   gin.H{"200": gin.H{"description": "OK"}},
				},
			},
			"/api/claims/{id}/moderate": gin.H{
				"put": gin.H{
					"tags":        []string{"Claims"},
					"summary":     "Complete or reject claim",
					"description": "Moderator only.",
					"security":    securityBearer,
					"responses": gin.H{
						"200": gin.H{"description": "OK"},
						"401": gin.H{"description": "Unauthorized"},
						"403": gin.H{"description": "Forbidden"},
					},
				},
			},
		},
	}
}

const swaggerUIHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>XRF Lab4 Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/swagger/openapi.json',
      dom_id: '#swagger-ui',
      deepLinking: true,
      persistAuthorization: true
    });
  </script>
</body>
</html>`
