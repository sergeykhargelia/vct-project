package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sergeykhargelia/vct-project/model"
	"github.com/sergeykhargelia/vct-project/templates"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("jwt"))

func (s *Server) RegisterPage(w http.ResponseWriter, r *http.Request) {
	templates.RegisterPage().Render(r.Context(), w)
}

func (s *Server) LoginPage(w http.ResponseWriter, r *http.Request) {
	templates.LoginPage().Render(r.Context(), w)
}

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	name := r.PostFormValue("name")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	if len(name) == 0 || len(email) == 0 || len(password) == 0 {
		templates.ErrorMessage("Fields should be non-empty").Render(r.Context(), w)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		templates.ErrorMessage("Failed to hash password").Render(r.Context(), w)
		return
	}

	user := model.User{Email: email, Name: name, PasswordHash: string(passwordHash)}
	if err := s.DB.Create(&user).Error; err != nil {
		templates.ErrorMessage("Error while creating user").Render(r.Context(), w)
		return
	}

	s.LoginHandler(w, r)
}

type Claims struct {
	UserID uint64
	jwt.RegisteredClaims
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	var user model.User
	if s.DB.Where("email = ?", email).First(&user).Error != nil {
		templates.ErrorMessage("User does not exist").Render(r.Context(), w)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		templates.ErrorMessage("Wrong password").Render(r.Context(), w)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		templates.ErrorMessage("Failed to sign token").Render(r.Context(), w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			templates.ErrorMessage("Error while reading cookie")
		}

		if cookie == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		tokenStr := cookie.Value

		var claims Claims
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("Signing algorithm mismatch")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			templates.ErrorMessage("Invalid token").Render(r.Context(), w)
			return
		}

		if claims.ExpiresAt.Time.Before(time.Now()) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
