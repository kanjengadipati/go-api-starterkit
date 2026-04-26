package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

type requestLogEntry struct {
	Level      string `json:"level"`
	Timestamp  string `json:"timestamp"`
	RequestID  string `json:"request_id,omitempty"`
	Method     string `json:"method,omitempty"`
	Path       string `json:"path,omitempty"`
	Status     int    `json:"status,omitempty"`
	LatencyMS  int64  `json:"latency_ms,omitempty"`
	ClientIP   string `json:"client_ip,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	UserID     any    `json:"user_id,omitempty"`
	ErrorCount int    `json:"error_count,omitempty"`
	Message    string `json:"message,omitempty"`
	Panic      any    `json:"panic,omitempty"`
}

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		entry := requestLogEntry{
			Level:      "info",
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			RequestID:  requestIDFromContext(c),
			Method:     c.Request.Method,
			Path:       path,
			Status:     c.Writer.Status(),
			LatencyMS:  time.Since(start).Milliseconds(),
			ClientIP:   c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			ErrorCount: len(c.Errors),
			Message:    "request completed",
		}

		if userID, exists := c.Get("user_id"); exists {
			entry.UserID = userID
		}

		if c.Writer.Status() >= http.StatusInternalServerError {
			entry.Level = "error"
		} else if c.Writer.Status() >= http.StatusBadRequest {
			entry.Level = "warn"
		}

		logStructured(entry)
	}
}

func RecoveryLogger() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		entry := requestLogEntry{
			Level:     "error",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: requestIDFromContext(c),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    http.StatusInternalServerError,
			ClientIP:  c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Message:   "panic recovered",
			Panic:     recovered,
		}

		if userID, exists := c.Get("user_id"); exists {
			entry.UserID = userID
		}

		logStructured(entry)
		httpx.Error(c, http.StatusInternalServerError, "internal server error")
	})
}

func requestIDFromContext(c *gin.Context) string {
	if value, exists := c.Get(RequestIDContextKey); exists {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}
	return ""
}

func logStructured(entry requestLogEntry) {
	payload, err := json.Marshal(entry)
	if err != nil {
		log.Printf(`{"level":"error","message":"failed to encode structured log","error":%q}`, err.Error())
		return
	}
	log.Print(string(payload))
}
