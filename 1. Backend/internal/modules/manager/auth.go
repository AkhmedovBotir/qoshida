package manager

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"qoshida/backend/internal/config"
	"qoshida/backend/internal/modules/eskiz"
	"qoshida/backend/internal/shared/apperror"
)

const (
	AccessCookie  = "qoshida_manager_access"
	RefreshCookie = "qoshida_manager_refresh"
	codeTTL       = 5 * time.Minute
	maxAttempts   = 5
	sendCooldown  = 60 * time.Second
	sendHourMax   = 5
)

type sendState struct {
	last time.Time
	hour []time.Time
}

type AuthService struct {
	repo *Repository
	sms  *eskiz.Service
	cfg  *config.Config
	mu   sync.Mutex
	gate map[string]*sendState
}

func NewAuthService(repo *Repository, sms *eskiz.Service, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, sms: sms, cfg: cfg, gate: map[string]*sendState{}}
}

func (s *AuthService) CheckPhone(ctx context.Context, phone string) (PhoneCheck, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return PhoneCheck{}, err
	}
	item, err := s.repo.FindByPhone(ctx, normalized)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return PhoneCheck{
				Exists:  false,
				Message: "Bu raqam admin tomonidan ro'yxatga olinmagan",
			}, nil
		}
		return PhoneCheck{}, apperror.Internal()
	}
	active := item.Status == StatusActive
	hasPassword := item.PasswordHash != ""
	needs := active && !hasPassword
	msg := "Parol kiritishingiz mumkin"
	if !active {
		msg = "Hisob faol emas"
	} else if needs {
		msg = "Parol o'rnatilmagan. SMS kod yuboriladi"
	}
	return PhoneCheck{
		Exists:      true,
		Active:      active,
		HasPassword: hasPassword,
		NeedsSetup:  needs,
		Message:     msg,
	}, nil
}

func (s *AuthService) SendCode(ctx context.Context, phone string) error {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return err
	}
	item, err := s.setupCandidate(ctx, normalized)
	if err != nil {
		return err
	}
	if err = s.allowSend(normalized); err != nil {
		return err
	}
	if s.sms == nil || !s.sms.Enabled() {
		return apperror.BadRequest("SMS xizmati sozlanmagan")
	}
	code := s.sms.NewCode()
	if err = s.repo.SaveCode(ctx, item.ID, normalized, PurposeSetup, code, codeTTL); err != nil {
		return apperror.Internal()
	}
	if _, err = s.sms.SendManagerPassword(ctx, normalized, code); err != nil {
		return apperror.BadRequest("SMS yuborilmadi. Keyinroq urinib ko'ring")
	}
	return nil
}

func (s *AuthService) VerifyCode(ctx context.Context, phone, code string) error {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return err
	}
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return apperror.BadRequest("Kod 6 xonali bo'lishi kerak")
	}
	id, _, storedHash, expiresAt, attempts, _, usedAt, err := s.repo.LatestCode(ctx, normalized, PurposeSetup)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.BadRequest("Avval SMS kod so'rang")
		}
		return apperror.Internal()
	}
	if usedAt != nil {
		return apperror.BadRequest("Kod allaqachon ishlatilgan")
	}
	if time.Now().After(expiresAt) {
		return apperror.New(http.StatusGone, "Kod muddati tugagan. Qayta yuboring")
	}
	if attempts >= maxAttempts {
		return apperror.New(http.StatusTooManyRequests, "Kod urinishlari tugadi. Qayta yuboring")
	}
	if storedHash != hashCode(normalized, PurposeSetup, code) {
		_ = s.repo.BumpAttempts(ctx, id)
		return apperror.BadRequest("Kod noto'g'ri")
	}
	if err = s.repo.MarkVerified(ctx, id); err != nil {
		return apperror.Internal()
	}
	return nil
}

func (s *AuthService) SetPassword(ctx context.Context, phone, password string) (Public, string, string, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return Public{}, "", "", err
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return Public{}, "", "", apperror.BadRequest("Parol kiritilishi shart")
	}
	if len(password) > 72 {
		return Public{}, "", "", apperror.BadRequest("Parol juda uzun")
	}
	item, err := s.setupCandidate(ctx, normalized)
	if err != nil {
		return Public{}, "", "", err
	}
	id, _, _, expiresAt, _, verifiedAt, usedAt, err := s.repo.LatestCode(ctx, normalized, PurposeSetup)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, "", "", apperror.BadRequest("Avval SMS kodni tasdiqlang")
		}
		return Public{}, "", "", apperror.Internal()
	}
	if usedAt != nil || verifiedAt == nil || time.Now().After(expiresAt) {
		return Public{}, "", "", apperror.BadRequest("SMS kod tasdiqlanmagan yoki muddati tugagan")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return Public{}, "", "", apperror.Internal()
	}
	item.PasswordHash = string(hash)
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, "", "", apperror.Internal()
	}
	_ = s.repo.MarkUsed(ctx, id)
	access, refresh, err := s.issue(ctx, updated)
	if err != nil {
		return Public{}, "", "", err
	}
	return updated.Public(), access, refresh, nil
}

func (s *AuthService) Login(ctx context.Context, phone, password string) (Public, string, string, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return Public{}, "", "", err
	}
	item, err := s.repo.FindByPhone(ctx, normalized)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, "", "", apperror.Unauthorized("Telefon yoki parol noto'g'ri")
		}
		return Public{}, "", "", apperror.Internal()
	}
	if item.Status != StatusActive {
		return Public{}, "", "", apperror.Forbidden("Hisob faol emas")
	}
	if item.PasswordHash == "" {
		return Public{}, "", "", apperror.Forbidden("Siz parol o'rnatishingiz kerak")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(item.PasswordHash), []byte(password)); err != nil {
		return Public{}, "", "", apperror.Unauthorized("Telefon yoki parol noto'g'ri")
	}
	access, refresh, err := s.issue(ctx, item)
	if err != nil {
		return Public{}, "", "", err
	}
	return item.Public(), access, refresh, nil
}

func (s *AuthService) Me(ctx context.Context, id uuid.UUID) (Public, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Public{}, apperror.Unauthorized("Sessiya yaroqsiz")
	}
	if item.Status != StatusActive {
		return Public{}, apperror.Forbidden("Hisob faol emas")
	}
	return item.Public(), nil
}

func (s *AuthService) Actor(ctx context.Context, id uuid.UUID) (Manager, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Manager{}, apperror.Unauthorized("Sessiya yaroqsiz")
	}
	if item.Status != StatusActive {
		return Manager{}, apperror.Forbidden("Hisob faol emas")
	}
	return item, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshRaw string) error {
	if refreshRaw == "" {
		return nil
	}
	return s.repo.RevokeRefresh(ctx, hashToken(refreshRaw))
}

func (s *AuthService) Refresh(ctx context.Context, refreshRaw string) (Public, string, string, error) {
	if strings.TrimSpace(refreshRaw) == "" {
		return Public{}, "", "", apperror.Unauthorized("Token talab qilinadi")
	}
	id, err := s.repo.FindValidRefresh(ctx, hashToken(refreshRaw))
	if err != nil {
		return Public{}, "", "", apperror.Unauthorized("Sessiya yaroqsiz")
	}
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Public{}, "", "", apperror.Unauthorized("Sessiya yaroqsiz")
	}
	if item.Status != StatusActive {
		return Public{}, "", "", apperror.Forbidden("Hisob faol emas")
	}
	_ = s.repo.RevokeRefresh(ctx, hashToken(refreshRaw))
	access, refresh, err := s.issue(ctx, item)
	if err != nil {
		return Public{}, "", "", err
	}
	return item.Public(), access, refresh, nil
}

func (s *AuthService) Restore(ctx context.Context, accessRaw, refreshRaw string) (uuid.UUID, string, string, error) {
	id, err := s.ParseAccess(accessRaw)
	if err == nil {
		return id, "", "", nil
	}
	item, access, refresh, err := s.Refresh(ctx, refreshRaw)
	if err != nil {
		return uuid.Nil, "", "", err
	}
	id, err = uuid.Parse(item.ID)
	if err != nil {
		return uuid.Nil, "", "", apperror.Unauthorized("Sessiya yaroqsiz")
	}
	return id, access, refresh, nil
}

func (s *AuthService) ParseAccess(token string) (uuid.UUID, error) {
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
	if use, _ := claims["use"].(string); use != "manager" {
		return uuid.Nil, apperror.Unauthorized("Token yaroqsiz")
	}
	sub, _ := claims["sub"].(string)
	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, apperror.Unauthorized("Token yaroqsiz")
	}
	return id, nil
}

func (s *AuthService) setupCandidate(ctx context.Context, phone string) (Manager, error) {
	item, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Manager{}, apperror.NotFound("Bu raqam admin tomonidan ro'yxatga olinmagan")
		}
		return Manager{}, apperror.Internal()
	}
	if item.Status != StatusActive {
		return Manager{}, apperror.Forbidden("Hisob faol emas")
	}
	if item.PasswordHash != "" {
		return Manager{}, apperror.Forbidden("Parol allaqachon o'rnatilgan. Kirishdan foydalaning")
	}
	return item, nil
}

func (s *AuthService) allowSend(phone string) error {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.gate[phone]
	if st == nil {
		st = &sendState{}
		s.gate[phone] = st
	}
	if !st.last.IsZero() && now.Sub(st.last) < sendCooldown {
		return apperror.New(http.StatusTooManyRequests, "Kodni qayta yuborish uchun 1 daqiqa kuting")
	}
	fresh := st.hour[:0]
	for _, t := range st.hour {
		if now.Sub(t) < time.Hour {
			fresh = append(fresh, t)
		}
	}
	if len(fresh) >= sendHourMax {
		return apperror.New(http.StatusTooManyRequests, "Soatiga 5 tadan ko'p SMS yuborib bo'lmaydi")
	}
	st.hour = append(fresh, now)
	st.last = now
	return nil
}

func (s *AuthService) issue(ctx context.Context, item Manager) (string, string, error) {
	now := time.Now()
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  item.ID.String(),
		"role": "manager_" + item.Type,
		"use":  "manager",
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
	if err = s.repo.StoreRefresh(ctx, item.ID, hashToken(refreshRaw), now.Add(s.cfg.JWTRefreshTTL)); err != nil {
		return "", "", apperror.Internal()
	}
	return accessToken, refreshRaw, nil
}

func (s *AuthService) SetCookies(w http.ResponseWriter, access, refresh string) {
	s.writeCookie(w, AccessCookie, access, s.cfg.JWTAccessTTL)
	s.writeCookie(w, RefreshCookie, refresh, s.cfg.JWTRefreshTTL)
}

func (s *AuthService) ClearCookies(w http.ResponseWriter) {
	s.writeCookie(w, AccessCookie, "", -time.Hour)
	s.writeCookie(w, RefreshCookie, "", -time.Hour)
}

func (s *AuthService) writeCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
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
