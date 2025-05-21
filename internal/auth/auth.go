package auth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// User представляет пользователя системы
type User struct {
	Username string
	Password string
}

// Middleware для базовой аутентификации
func BasicAuth(username, password string, enabled bool) gin.HandlerFunc {
	// Если аутентификация выключена, просто пропускаем запросы
	if !enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Получение заголовка авторизации
		auth := c.GetHeader("Authorization")

		// Проверка наличия заголовка
		if auth == "" {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Проверка формата
		if !strings.HasPrefix(auth, "Basic ") {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Извлечение закодированных учетных данных
		encoded := strings.TrimPrefix(auth, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Разбираем декодированные учетные данные
		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Проверяем учетные данные
		if credentials[0] != username || credentials[1] != password {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Сохраняем информацию о пользователе в контексте
		c.Set("user", &User{
			Username: username,
			Password: password,
		})

		// Если все проверки пройдены
		c.Next()
	}
}

// GetCurrentUser возвращает текущего аутентифицированного пользователя
func GetCurrentUser(c *gin.Context) *User {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil
	}

	user, ok := userInterface.(*User)
	if !ok {
		return nil
	}

	return user
}

// IsAuthenticated проверяет, аутентифицирован ли пользователь
func IsAuthenticated(c *gin.Context) bool {
	return GetCurrentUser(c) != nil
}

// RequireAuth - middleware, требующее аутентификации
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAuthenticated(c) {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
