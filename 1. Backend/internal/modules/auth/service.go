package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"qoshida/backend/internal/config"
	"qoshida/backend/internal/modules/admin"
	"qoshida/backend/internal/shared/apperror"
)

const (
	AccessCookie  = "qoshida_admin_access"
	RefreshCookie = "qoshida_admin_refresh"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Service struct {
	admins *admin.Repository
	tokens *TokenRepository
	cfg    *config.Config
}

func NewService(admins *admin.Repository, tokens *TokenRepository, cfg *config.Config) *Service {
	return &Service{admins: admins, tokens: tokens, cfg: cfg}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (admin.Public, string, string, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || len(req.Password) < 8 {
		return admin.Public{}, "", "", apperror.Unauthorized("Username yoki parol noto'g'ri")
	}

	user, err := s.admins.FindByUsername(ctx, username)
	if err != nil {
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$12$R9h/cIPz0gi.URNNX3kh2OPST9/PgBkqquzi.Ss7KIUgO2t0jWMUW"), []byte(req.Password))
		return admin.Public{}, "", "", apperror.Unauthorized("Username yoki parol noto'g'ri")
	}
	if !user.IsActive {
		return admin.Public{}, "", "", apperror.Forbidden("Hisob faol emas")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return admin.Public{}, "", "", apperror.Unauthorized("Username yoki parol noto'g'ri")
	}

	access, refresh, err := s.issue(ctx, user)
	if err != nil {
		return admin.Public{}, "", "", err
	}
	return user.Public(), access, refresh, nil
}

func (s *Service) Me(ctx context.Context, id uuid.UUID) (admin.Public, error) {
	user, err := s.admins.FindByID(ctx, id)
	if err != nil {
		return admin.Public{}, apperror.Unauthorized("Sessiya yaroqsiz")
	}
	if !user.IsActive {
		return admin.Public{}, apperror.Forbidden("Hisob faol emas")
	}
	return user.Public(), nil
}

func (s *Service) Logout(ctx context.Context, refreshRaw string) error {
	if refreshRaw == "" {
		return nil
	}
	return s.tokens.Revoke(ctx, hashToken(refreshRaw))
}

func (s *Service) Refresh(ctx context.Context, refreshRaw string) (admin.Public, string, string, error) {
	if strings.TrimSpace(refreshRaw) == "" {
		return admin.Public{}, "", "", apperror.Unauthorized("Token talab qilinadi")
	}
	adminID, err := s.tokens.FindValid(ctx, hashToken(refreshRaw))
	if err != nil {
		return admin.Public{}, "", "", apperror.Unauthorized("Sessiya yaroqsiz")
	}
	user, err := s.admins.FindByID(ctx, adminID)
	if err != nil {
		return admin.Public{}, "", "", apperror.Unauthorized("Sessiya yaroqsiz")
	}
	if !user.IsActive {
		return admin.Public{}, "", "", apperror.Forbidden("Hisob faol emas")
	}
	_ = s.tokens.Revoke(ctx, hashToken(refreshRaw))
	access, refresh, err := s.issue(ctx, user)
	if err != nil {
		return admin.Public{}, "", "", err
	}
	return user.Public(), access, refresh, nil
}

func (s *Service) Restore(ctx context.Context, accessRaw, refreshRaw string) (uuid.UUID, string, string, error) {
	id, err := s.ParseAccess(accessRaw)
	if err == nil {
		return id, "", "", nil
	}
	user, access, refresh, err := s.Refresh(ctx, refreshRaw)
	if err != nil {
		return uuid.Nil, "", "", err
	}
	id, err = uuid.Parse(user.ID)
	if err != nil {
		return uuid.Nil, "", "", apperror.Unauthorized("Sessiya yaroqsiz")
	}
	return id, access, refresh, nil
}

func (s *Service) ParseAccess(token string) (uuid.UUID, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, apperror.Unauthorized("Token yaroqsiz")
		}
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithIssuer(s.cfg.JWTIssuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !parsed.Valid {
		return uuid.Nil, apperror.Unauthorized("Token yaroqsiz")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, apperror.Unauthorized("Token yaroqsiz")
	}
	if typ, _ := claims["typ"].(string); typ != "access" {
		return uuid.Nil, apperror.Unauthorized("Token yaroqsiz")
	}
	sub, _ := claims["sub"].(string)
	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, apperror.Unauthorized("Token yaroqsiz")
	}
	return id, nil
}

func (s *Service) issue(ctx context.Context, user admin.Admin) (string, string, error) {
	now := time.Now()
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": user.Role,
		"typ":  "access",
		"iss":  s.cfg.JWTIssuer,
		"iat":  now.Unix(),
		"exp":  now.Add(s.cfg.JWTAccessTTL).Unix(),
	})
	accessToken, err := access.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", apperror.Internal()
	}

	refreshRaw, err := randomToken()
	if err != nil {
		return "", "", apperror.Internal()
	}
	if err = s.tokens.Store(ctx, user.ID, hashToken(refreshRaw), now.Add(s.cfg.JWTRefreshTTL)); err != nil {
		return "", "", apperror.Internal()
	}
	return accessToken, refreshRaw, nil
}

func (s *Service) SetCookies(w http.ResponseWriter, access, refresh string) {
	s.writeCookie(w, AccessCookie, access, s.cfg.JWTAccessTTL)
	s.writeCookie(w, RefreshCookie, refresh, s.cfg.JWTRefreshTTL)
}

func (s *Service) ClearCookies(w http.ResponseWriter) {
	s.writeCookie(w, AccessCookie, "", -time.Hour)
	s.writeCookie(w, RefreshCookie, "", -time.Hour)
}

func (s *Service) writeCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func AccessFromRequest(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	cookie, err := r.Cookie(AccessCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func RefreshFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(RefreshCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}
