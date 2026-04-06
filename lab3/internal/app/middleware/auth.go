package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"xrfApp/internal/app/auth"
	"xrfApp/internal/app/session"
)

const contextAuthUserKey = "auth_user"

type AuthUser struct {
	ID    uint
	Login string
	Role  string
	Token string
}

func RequireAuth(tokens *auth.Manager, tokenStore *session.Manager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if tokens == nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "jwt manager is not configured",
			})
			return
		}

		rawToken, err := auth.ExtractBearerToken(ctx.GetHeader("Authorization"))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		claims, err := tokens.ParseToken(rawToken)
		if err != nil {
			code := http.StatusUnauthorized
			if err == auth.ErrExpiredToken {
				code = http.StatusUnauthorized
			}

			ctx.AbortWithStatusJSON(code, gin.H{
				"message": err.Error(),
			})
			return
		}

		if tokenStore != nil {
			tokenRecord, err := tokenStore.GetToken(ctx.Request.Context(), rawToken)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "token check failed",
				})
				return
			}
			if tokenRecord == nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "token is not found or expired",
				})
				return
			}

			if tokenRecord.UserID != claims.UserID ||
				!strings.EqualFold(strings.TrimSpace(tokenRecord.Login), strings.TrimSpace(claims.Login)) ||
				!strings.EqualFold(strings.TrimSpace(tokenRecord.Role), strings.TrimSpace(claims.Role)) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "token record does not match token payload",
				})
				return
			}
		}

		ctx.Set(contextAuthUserKey, AuthUser{
			ID:    claims.UserID,
			Login: claims.Login,
			Role:  strings.ToLower(strings.TrimSpace(claims.Role)),
			Token: rawToken,
		})
		ctx.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		clean := strings.ToLower(strings.TrimSpace(role))
		if clean != "" {
			allowed[clean] = struct{}{}
		}
	}

	return func(ctx *gin.Context) {
		user, ok := CurrentUser(ctx)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "authorization required",
			})
			return
		}

		if _, exists := allowed[strings.ToLower(strings.TrimSpace(user.Role))]; !exists {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "forbidden: insufficient permissions",
			})
			return
		}

		ctx.Next()
	}
}

func CurrentUser(ctx *gin.Context) (AuthUser, bool) {
	value, exists := ctx.Get(contextAuthUserKey)
	if !exists {
		return AuthUser{}, false
	}

	user, ok := value.(AuthUser)
	if !ok || user.ID == 0 {
		return AuthUser{}, false
	}
	return user, true
}
