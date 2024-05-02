package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vladyslavpavlenko/peparesu/internal/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

// signupBody is the signup request body structure.
type signupBody struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// loginBody is the login request body structure.
type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignUp handles the signup request.
func (m *Repository) SignUp(w http.ResponseWriter, r *http.Request) {
	var body signupBody

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&body)
	defer r.Body.Close()
	if err != nil {
		_ = m.errorJSON(w, errors.New("failed to read body"))
		return
	}

	if body.FirstName == "" || body.LastName == "" || !validateEmail(body.Email) || len(body.Password) < 8 {
		_ = m.errorJSON(w, errors.New("bad credentials provided"))
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		_ = m.errorJSON(w, fmt.Errorf("error hashing password: %v", err))
		return
	}

	// Create a new user model
	user := models.User{
		FirstName:  body.FirstName,
		LastName:   body.LastName,
		Email:      body.Email,
		Password:   string(hashedPassword),
		UserTypeID: 1, // User
	}

	// Add user to the database
	result := m.App.DB.Create(&user)
	if result.Error != nil {
		_ = m.errorJSON(w, fmt.Errorf("error creating user: %v", result.Error))
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "user created",
	}

	_ = m.writeJSON(w, http.StatusCreated, payload)
}

// Login handles the login request.
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody

	// Get the email and password off request body
	err := json.NewDecoder(r.Body).Decode(&body)
	defer r.Body.Close()
	if err != nil {
		_ = m.errorJSON(w, errors.New("failed to read body"))
		return
	}

	// Look up the requested user
	var user models.User
	m.App.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		_ = m.errorJSON(w, errors.New("invalid email"), http.StatusUnauthorized)
		return
	}

	// Compare sent in password with saved user password hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		_ = m.errorJSON(w, errors.New("invalid password"), http.StatusUnauthorized)
		return
	}

	// Generate a JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 14).Unix(),
	})

	// Sign in and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(m.App.Env.JWTSecret))
	if err != nil {
		_ = m.errorJSON(w, errors.New("failed to create JWT token"))
		return
	}

	// create and set a cookie
	cookie := http.Cookie{
		Name:     "user_jwt",
		Path:     "/",
		Value:    tokenString,
		MaxAge:   3600 * 24 * 30,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("logged in as %s %s", user.FirstName, user.LastName),
	}

	_ = m.writeJSON(w, http.StatusOK, payload)
}

// Logout handles the logout request by invalidating the user's session cookie.
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "user_jwt",
		Path:     "/",
		Value:    "",
		MaxAge:   -1,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	payload := jsonResponse{
		Error:   false,
		Message: "logged out",
	}

	_ = m.writeJSON(w, http.StatusOK, payload)
}
