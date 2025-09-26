package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const maxLogBody = 1 << 20 // 1 MB

func DebugRequestLogger() gin.HandlerFunc {
	isSensitive := func(k string) bool {
		k = strings.ToLower(k)
		return k == "authorization" || k == "cookie" || k == "set-cookie" || k == "proxy-authorization"
	}

	return func(c *gin.Context) {
		start := time.Now()

		headers := make(map[string]string, len(c.Request.Header))
		for k, v := range c.Request.Header {
			if isSensitive(k) {
				headers[k] = "[REDACTED]"
			} else {
				headers[k] = strings.Join(v, ",")
			}
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			data, err := io.ReadAll(c.Request.Body)
			if err == nil {
				bodyBytes = data
			} else {
				logrus.Debugf("[GIN DEBUG] failed to read request body: %v", err)
			}
			// восстанавливаем тело для последующих обработчиков
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		ctype := c.GetHeader("Content-Type")

		var bodyForLog string
		switch {
		case len(bodyBytes) == 0:
			bodyForLog = "<empty>"

		case strings.HasPrefix(ctype, "application/json"):
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, bodyBytes, "", "  "); err == nil {
				s := pretty.String()
				if len(s) > maxLogBody {
					bodyForLog = s[:maxLogBody] + "\n...(truncated)"
				} else {
					bodyForLog = s
				}
			} else {
				if len(bodyBytes) > maxLogBody {
					bodyForLog = string(bodyBytes[:maxLogBody]) + "\n...(truncated)"
				} else {
					bodyForLog = string(bodyBytes)
				}
			}

		case strings.HasPrefix(ctype, "text/") || strings.Contains(ctype, "form-urlencoded"):
			if len(bodyBytes) > maxLogBody {
				bodyForLog = string(bodyBytes[:maxLogBody]) + "\n...(truncated)"
			} else {
				bodyForLog = string(bodyBytes)
			}

		default:
			bodyForLog = "<not-logged: content-type=" + ctype + " length=" + strconv.Itoa(len(bodyBytes)) + ">"
		}

		// Пропускаем обработчики
		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		logrus.Debugf("[GIN DEBUG] %s %s -> %d (%s)\nHeaders: %v\nBody:\n%s\n",
			c.Request.Method, c.Request.URL.RequestURI(), status, latency, headers, bodyForLog)
	}
}
