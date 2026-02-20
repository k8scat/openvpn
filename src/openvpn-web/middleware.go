package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func customLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}
		return fmt.Sprintf("[OPENVPN-WEB] %v GIN |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006-01-02 15:04:05.000"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)
	})
}

func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		if c.Request.URL.Path == "/ovpn/login" || c.Request.URL.Path == "/ovpn/history" {
			if c.ClientIP() == "127.0.0.1" || c.ClientIP() == "::1" {
				c.Next()
				return
			}
		}

		if user == nil {
			c.Redirect(302, "/login")
			c.Abort()
			return
		}

		if user, ok := user.(string); ok {
			if c.Request.URL.Path != "/" && !strings.HasPrefix(c.Request.URL.Path, "/client") && user != adminUsername {
				c.Redirect(302, "/")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
