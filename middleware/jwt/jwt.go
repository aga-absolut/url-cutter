package jwt

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/golang-jwt/jwt/v4"
)

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 3)),
		},
		UserID: config.UserID,
	})

	tokenString, err := token.SignedString(config.SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserID(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return config.SecretKey, nil
	})
	if err != nil {
		return -1
	}
	if !token.Valid {
		return -1
	}
	return claims.UserID
}

func AuthMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		if err != nil {
			token, _ := BuildJWTString()
			c = &http.Cookie{
				Name:     "token",
				Value:    token,
				HttpOnly: true,
				Path:     "/",
			}
			http.SetCookie(w, c)
			r.AddCookie(c)
		}
		h(w, r)
	})
}
