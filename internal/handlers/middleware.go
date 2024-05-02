package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vladyslavpavlenko/peparesu/internal/models"
	"net/http"
	"time"
)

// RequireAuth is a middleware that checks for the presence and validity of a JWT in the request cookie.
func (m *Repository) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the cookie of request
		tokenCookie, err := r.Cookie("user_jwt")
		if err != nil {
			_ = m.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)

			return
		}

		// Decode/validate it
		token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(m.App.Env.JWTSecret), nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check the exp
			if float64(time.Now().Unix()) > claims["exp"].(float64) {
				_ = m.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// Find the user with token sub
			var user models.User

			m.App.DB.Preload("UserType").First(&user, "id = ?", claims["sub"])

			if user.ID == 0 {
				_ = m.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user", user)

			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			_ = m.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
			return
		}
	})
}

// RequireNoAuth is a middleware that checks for the presence of a JWT in the request cookie.
func (m *Repository) RequireNoAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to get the cookie from the request
		tokenCookie, err := r.Cookie("user_jwt")

		// If no cookie, proceed with the next handler
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Attempt to decode and validate the JWT
		token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.App.Env.JWTSecret), nil
		})

		if err == nil && token.Valid {
			_ = m.errorJSON(w, errors.New("already authenticated"), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
