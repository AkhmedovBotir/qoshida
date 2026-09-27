package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"qoshida/backend/internal/config"
)

func Register(r *chi.Mux, cfg *config.Config) {
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(timeoutExceptImport(cfg.RequestTimeout))
	r.Use(securityHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(httprate.LimitByIP(cfg.RateLimitRPM, time.Minute))
	r.Use(maxBytes(cfg.BodyLimitBytes))
	r.Use(jsonOnly)
}

func timeoutExceptImport(d time.Duration) func(http.Handler) http.Handler {
	withTimeout := chimw.Timeout(d)
	return func(next http.Handler) http.Handler {
		timed := withTimeout(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isLargeWrite(r) {
				next.ServeHTTP(w, r)
				return
			}
			timed.ServeHTTP(w, r)
		})
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		w.Header().Set("Cache-Control", "no-store")
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func maxBytes(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				bodyLimit := limit
				if isProductWrite(r) || isIdentWrite(r) || isServiceWrite(r) || isCustomerWrite(r) || isCategoryWrite(r) {
					bodyLimit = 100 << 20
				}
				r.Body = http.MaxBytesReader(w, r.Body, bodyLimit)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isLargeWrite(r *http.Request) bool {
	if r.Method == http.MethodPost && (strings.HasSuffix(r.URL.Path, "/regions/import") || strings.HasSuffix(r.URL.Path, "/categories/import") || strings.HasSuffix(r.URL.Path, "/activity-types/import")) {
		return true
	}
	return isProductWrite(r) || isIdentWrite(r) || isServiceWrite(r) || isCustomerWrite(r) || isCategoryWrite(r)
}

func isCategoryWrite(r *http.Request) bool {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return false
	}
	if strings.HasSuffix(r.URL.Path, "/import") {
		return false
	}
	return strings.Contains(r.URL.Path, "/categories")
}

func isIdentWrite(r *http.Request) bool {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return false
	}
	return strings.Contains(r.URL.Path, "/identification")
}

func isProductWrite(r *http.Request) bool {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return false
	}
	path := r.URL.Path
	if strings.Contains(path, "/shop-templates") {
		return true
	}
	if !strings.Contains(path, "/products") {
		return false
	}
	return !strings.HasSuffix(path, "/approve") && !strings.HasSuffix(path, "/reject")
}

func isServiceWrite(r *http.Request) bool {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return false
	}
	path := r.URL.Path
	if strings.HasSuffix(path, "/approve") || strings.HasSuffix(path, "/reject") {
		return false
	}
	return strings.Contains(path, "/provider-services") || strings.Contains(path, "/services")
}

func isCustomerWrite(r *http.Request) bool {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		return false
	}
	return strings.Contains(r.URL.Path, "/mijoz/profile")
}

func jsonOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := strings.ToLower(r.Header.Get("Content-Type"))
			if !strings.HasPrefix(ct, "application/json") {
				http.Error(w, `{"success":false,"error":"Content-Type application/json bo'lishi kerak"}`, http.StatusUnsupportedMediaType)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
