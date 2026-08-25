package middleware

import "net/http"

// CORSMiddleware จัดการ Cross-Origin Resource Sharing สำหรับ Web (Next.js) & Mobile (React Native)
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// อนุญาตให้เรียกจาก Origin ภายนอก (Next.js, Mobile App, Postman)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")

		// ดักจับ Preflight Request (OPTIONS) ที่เบราว์เซอร์จะยิงมาถามก่อนเสมอ
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
