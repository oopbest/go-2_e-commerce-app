package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery Middleware ดักจับ Panic เพื่อไม่ให้ Server พัง และตอบกลับ 500 อย่างสุภาพ
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// บันทึก Error Log พร้อม Stack Trace ของโค้ดที่เกิดปัญหา
				slog.Error("CRITICAL: Panic Recovered",
					"error", err,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)

				sendError(w, http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
