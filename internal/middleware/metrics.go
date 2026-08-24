package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/oopbest/ecommerce-app/internal/metrics"
)

// Regex สำหรับแปลงตัวเลข ID ใน Path ให้เป็น {id} เพื่อป้องกัน High Cardinality
var idRegex = regexp.MustCompile(`/\d+`)

func normalizePath(path string) string {
	return idRegex.ReplaceAllString(path, "/{id}")
}

// MetricsMiddleware ดักจับ Request เพื่อบันทึก Counter, Histogram, และ Gauge
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ข้ามการบันทึก Metrics ของตัว Endpoint /metrics เอง
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// 1. เพิ่มเกจวัด In-Flight (+1) และลดลงเมื่อเสร็จสิ้น (-1)
		metrics.HttpRequestsInFlight.Inc()
		defer metrics.HttpRequestsInFlight.Dec()

		start := time.Now()
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 2. ส่งต่อให้ Handler ทำงาน
		next.ServeHTTP(wrapped, r)

		// 3. คำนวณเวลาที่ใช้ และ Normalize Path
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(wrapped.statusCode)
		path := normalizePath(r.URL.Path)

		// 4. บันทึกค่าลง Prometheus Metrics
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, path, statusCode).Inc()
		metrics.HttpRequestDuration.WithLabelValues(r.Method, path, statusCode).Observe(duration)
	})
}
