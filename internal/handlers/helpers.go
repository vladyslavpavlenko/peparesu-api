package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vladyslavpavlenko/peparesu/internal/models"
	"io"
	"net/http"
	"net/mail"
)

func (m *Repository) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // 1 MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	return nil
}

func (m *Repository) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (m *Repository) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return m.writeJSON(w, statusCode, payload)
}

// validateEmail validates the email address.
func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// getUserFromToken extracts the user ID from the JWT token in the request cookie.
func (m *Repository) getUserFromToken(r *http.Request) (uint, error) {
	cookie, err := r.Cookie("user_jwt")
	if err != nil {
		return 0, errors.New("no token found")
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.App.Env.JWTSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId, ok := claims["sub"].(float64)
		if !ok {
			return 0, errors.New("invalid token claims")
		}
		return uint(userId), nil
	}

	return 0, errors.New("invalid token")
}

// isAdmin checks if the given user ID corresponds to an admin user.
func (m *Repository) isAdmin(userID uint) bool {
	var user models.User
	if err := m.App.DB.First(&user, "id = ?", userID).Error; err != nil {
		return false
	}
	return user.UserTypeID == 2
}
