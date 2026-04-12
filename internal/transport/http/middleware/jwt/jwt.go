package jwt

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/aga-absolut/url-cutter/internal/config"
)

var UserID int

// Claims структура.
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// BuildJWTString генерирует JWT токен.
func BuildJWTString() (string, error) {
	UserID++
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenExpTime)),
		},
		UserID: UserID,
	})

	tokenString, err := token.SignedString(config.SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserID извлекает ID пользователя с помощью JWT токена.
func GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return config.SecretKey, nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, err
	}
	return claims.UserID, nil
}

// AuthMiddleware проверяет пользователя на авторизованность и создает новую куки если ее нет.
func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("token")
		if err != nil {
			token, err := BuildJWTString()
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			c := &http.Cookie{
				Name:     "token",
				Value:    token,
				HttpOnly: true,
				Path:     "/",
			}
			http.SetCookie(w, c)
			r.AddCookie(c)
		}

		h.ServeHTTP(w, r)
	})
}
