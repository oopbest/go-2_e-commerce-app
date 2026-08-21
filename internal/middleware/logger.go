package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter struct ครอบ http.ResponseWriter เพื่อแอบดักจับ Status Code ที่ส่งกลับไป
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger Middleware บันทึกรายละเอียดของทุก HTTP Request ด้วย slog
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter เพื่อดักจับ Status Code (ค่าเริ่มต้น 200)
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// ปล่อยให้ Request ทำงานต่อไป
		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)

		// บันทึก Structured Log
		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrappedWriter.statusCode,
			"duration_ms", duration.Milliseconds(),
			"ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}
