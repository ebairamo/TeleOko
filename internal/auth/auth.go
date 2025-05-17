package auth

import (
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
	// TODO: Реализовать middleware для базовой аутентификации
	// 1. Если аутентификация выключена, просто пропускать запросы
	// 2. Проверять заголовок Authorization
	// 3. Декодировать и проверять учетные данные
	// 4. При успешной аутентификации пропускать запрос
	// 5. При неуспешной - возвращать 401 Unauthorized

	if !enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// TODO: Реализовать проверку аутентификации

		// Получение заголовка авторизации
		auth := c.GetHeader("Authorization")

		// Проверка наличия заголовка
		if auth == "" {
			c.Header("WWW-Authenticate", "Basic realm=Restricted")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Проверка формата
		if !strings.HasPrefix(auth, "Basic ") {
			c.Header("WWW-Authenticate", "Basic realm=Restricted")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// TODO: Декодирование и проверка учетных данных

		// Если все проверки пройдены
		c.Next()
	}
}
