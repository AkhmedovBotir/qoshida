package customer

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"qoshida/backend/internal/config"
	"qoshida/backend/internal/modules/eskiz"
	"qoshida/backend/internal/shared/apperror"
)

const (
	AccessCookie  = "qoshida_customer_access"
	RefreshCookie = "qoshida_customer_refresh"
	codeTTL       = 5 * time.Minute
	maxAttempts   = 5
	sendCooldown  = 60 * time.Second
	sendHourMax   = 5
)

var phoneRE = regexp.MustCompile(`^\+998[0-9]{9}$`)

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
				Exists:      false,
				CanRegister: true,
				Message:     "Ro‘yxatdan o‘tishingiz mumkin",
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
		msg = "Parol o‘rnatilmagan. SMS kod yuboriladi"
	}
	return PhoneCheck{
		Exists:      true,
		Active:      active,
		HasPassword: hasPassword,
		NeedsSetup:  needs,
		Message:     msg,
	}, nil
}

func (s *AuthService) SendCode(ctx context.Context, phone, purpose string) (SendCodeResult, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return SendCodeResult{}, err
	}
	item, err := s.repo.FindByPhone(ctx, normalized)
	exists := err == nil
	if err != nil && !errors.Is(err, ErrNotFound) {
		return SendCodeResult{}, apperror.Internal()
	}
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		if exists {
			purpose = PurposeLogin
		} else {
			purpose = PurposeRegister
		}
	} else {
		purpose, err = normalizePurpose(purpose)
		if err != nil {
			return SendCodeResult{}, err
		}
	}
	var customerID *uuid.UUID
	switch purpose {
	case PurposeRegister:
		if exists {
			return SendCodeResult{}, apperror.Conflict("Bu raqam allaqachon ro‘yxatdan o‘tgan")
		}
	case PurposeLogin, PurposeSetup, PurposeReset:
		if !exists {
			return SendCodeResult{}, apperror.NotFound("Bu raqam ro‘yxatga olinmagan")
		}
		if item.Status != StatusActive {
			return SendCodeResult{}, apperror.Forbidden("Hisob faol emas")
		}
		if purpose == PurposeSetup && item.PasswordHash != "" {
			return SendCodeResult{}, apperror.Forbidden("Parol allaqachon o‘rnatilgan. Kirishdan foydalaning")
		}
		customerID = &item.ID
	}
	if err = s.allowSend(normalized); err != nil {
		return SendCodeResult{}, err
	}
	if s.sms == nil || !s.sms.Enabled() {
		return SendCodeResult{}, apperror.BadRequest("SMS xizmati sozlanmagan")
	}
	code := s.sms.NewCode()
	if err = s.repo.SaveCode(ctx, customerID, normalized, purpose, code, codeTTL); err != nil {
		return SendCodeResult{}, apperror.Internal()
	}
	var sendErr error
	switch purpose {
	case PurposeRegister:
		_, sendErr = s.sms.SendMarketplaceRegister(ctx, normalized, code)
	case PurposeReset:
		_, sendErr = s.sms.SendMarketplaceReset(ctx, normalized, code)
	default:
		_, sendErr = s.sms.SendMarketplaceLogin(ctx, normalized, code)
	}
	if sendErr != nil {
		return SendCodeResult{}, apperror.BadRequest("SMS yuborilmadi. Keyinroq urinib ko‘ring")
	}
	msg := "SMS kodi yuborildi. Kod 5 daqiqa amal qiladi"
	return SendCodeResult{Status: "sent", Message: msg, Purpose: purpose, Exists: exists}, nil
}

func (s *AuthService) VerifyCode(ctx context.Context, phone, code, purpose string) (VerifyResult, string, string, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return VerifyResult{}, "", "", err
	}
	purpose, err = normalizePurpose(purpose)
	if err != nil {
		return VerifyResult{}, "", "", err
	}
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return VerifyResult{}, "", "", apperror.BadRequest("Kod 6 xonali bo‘lishi kerak")
	}
	id, _, storedHash, expiresAt, attempts, _, usedAt, err := s.repo.LatestCode(ctx, normalized, purpose)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return VerifyResult{}, "", "", apperror.BadRequest("Avval SMS kod so‘rang")
		}
		return VerifyResult{}, "", "", apperror.Internal()
	}
	if usedAt != nil {
		return VerifyResult{}, "", "", apperror.BadRequest("Kod allaqachon ishlatilgan")
	}
	if time.Now().After(expiresAt) {
		return VerifyResult{}, "", "", apperror.BadRequest("Kod muddati tugagan")
	}
	if attempts >= maxAttempts {
		return VerifyResult{}, "", "", apperror.Forbidden("Urinishlar soni tugadi. Qayta kod so‘rang")
	}
	if storedHash != hashCode(normalized, purpose, code) {
		_ = s.repo.BumpAttempts(ctx, id)
		return VerifyResult{}, "", "", apperror.BadRequest("Kod noto‘g‘ri")
	}
	if err = s.repo.MarkVerified(ctx, id); err != nil {
		return VerifyResult{}, "", "", apperror.Internal()
	}

	out := VerifyResult{Status: "verified", Purpose: purpose, Exists: purpose != PurposeRegister}
	if purpose == PurposeRegister {
		return out, "", "", nil
	}

	item, err := s.repo.FindByPhone(ctx, normalized)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return VerifyResult{}, "", "", apperror.NotFound("Bu raqam ro‘yxatga olinmagan")
		}
		return VerifyResult{}, "", "", apperror.Internal()
	}
	if item.Status != StatusActive {
		return VerifyResult{}, "", "", apperror.Forbidden("Hisob faol emas")
	}
	_ = s.repo.MarkUsed(ctx, id)
	access, refresh, err := s.issue(ctx, item)
	if err != nil {
		return VerifyResult{}, "", "", err
	}
	pub := item.Public()
	out.Customer = &pub
	out.Exists = true
	return out, access, refresh, nil
}

func (s *AuthService) Register(ctx context.Context, phone string, fields ProfileFields) (Customer, string, string, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return Customer{}, "", "", err
	}
	if err = s.requireVerified(ctx, normalized, PurposeRegister); err != nil {
		return Customer{}, "", "", err
	}
	if _, err = s.repo.FindByPhone(ctx, normalized); err == nil {
		return Customer{}, "", "", apperror.Conflict("Bu raqam allaqachon ro‘yxatdan o‘tgan")
	} else if !errors.Is(err, ErrNotFound) {
		return Customer{}, "", "", apperror.Internal()
	}
	item := Customer{
		ID:         uuid.New(),
		Name:       fields.Name,
		FirstName:  fields.FirstName,
		LastName:   fields.LastName,
		BirthDate:  fields.BirthDate,
		Phone:      normalized,
		RegionID:   fields.RegionID,
		DistrictID: fields.DistrictID,
		MFYID:      fields.MFYID,
		Address:    fields.Address,
		Status:     StatusActive,
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Customer{}, "", "", apperror.Conflict("Bu raqam allaqachon ro‘yxatdan o‘tgan")
		}
		return Customer{}, "", "", apperror.Internal()
	}
	id, _, _, _, _, _, _, err := s.repo.LatestCode(ctx, normalized, PurposeRegister)
	if err == nil {
		_ = s.repo.MarkUsed(ctx, id)
	}
	access, refresh, err := s.issue(ctx, created)
	if err != nil {
		return Customer{}, "", "", err
	}
	return created, access, refresh, nil
}

func (s *AuthService) SetPassword(ctx context.Context, phone, password, purpose string) (Customer, string, string, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return Customer{}, "", "", err
	}
	purpose, err = normalizePurpose(purpose)
	if err != nil {
		return Customer{}, "", "", err
	}
	if purpose != PurposeSetup && purpose != PurposeReset {
		return Customer{}, "", "", apperror.BadRequest("Maqsad noto‘g‘ri")
	}
	if err = simplePassword(password); err != nil {
		return Customer{}, "", "", err
	}
	if err = s.requireVerified(ctx, normalized, purpose); err != nil {
		return Customer{}, "", "", err
	}
	item, err := s.repo.FindByPhone(ctx, normalized)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Customer{}, "", "", apperror.NotFound("Bu raqam ro‘yxatga olinmagan")
		}
		return Customer{}, "", "", apperror.Internal()
	}
	if item.Status != StatusActive {
		return Customer{}, "", "", apperror.Forbidden("Hisob faol emas")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Customer{}, "", "", apperror.Internal()
	}
	item.PasswordHash = string(hash)
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Customer{}, "", "", apperror.Internal()
	}
	id, _, _, _, _, _, _, err := s.repo.LatestCode(ctx, normalized, purpose)
	if err == nil {
		_ = s.repo.MarkUsed(ctx, id)
	}
	access, refresh, err := s.issue(ctx, updated)
	if err != nil {
		return Customer{}, "", "", err
	}
	return updated, access, refresh, nil
}

func (s *AuthService) Login(ctx context.Context, phone, password string) (Customer, string, string, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return Customer{}, "", "", err
	}
	item, err := s.repo.FindByPhone(ctx, normalized)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Customer{}, "", "", apperror.Unauthorized("Telefon yoki parol noto‘g‘ri")
		}
		return Customer{}, "", "", apperror.Internal()
	}
	if item.Status != StatusActive {
		return Customer{}, "", "", apperror.Forbidden("Hisob faol emas")
	}
	if item.PasswordHash == "" {
		return Customer{}, "", "", apperror.Forbidden("Avval parol o‘rnating")
	}
	if bcrypt.CompareHashAndPassword([]byte(item.PasswordHash), []byte(password)) != nil {
		return Customer{}, "", "", apperror.Unauthorized("Telefon yoki parol noto‘g‘ri")
	}
	access, refresh, err := s.issue(ctx, item)
	if err != nil {
		return Customer{}, "", "", err
	}
	return item, access, refresh, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshRaw string) error {
	if strings.TrimSpace(refreshRaw) == "" {
		return nil
	}
	return s.repo.RevokeRefresh(ctx, hashToken(refreshRaw))
}

func (s *AuthService) Refresh(ctx context.Context, refreshRaw string) (Public, string, string, error) {
	if strings.TrimSpace(refreshRaw) == "" {
		return Public{}, "", "", apperror.Unauthorized("Sessiya yaroqsiz")
	}
	id, err := s.repo.FindValidRefresh(ctx, hashToken(refreshRaw))
	if err != nil {
		return Public{}, "", "", apperror.Unauthorized("Sessiya yaroqsiz")
	}
	_ = s.repo.RevokeRefresh(ctx, hashToken(refreshRaw))
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Public{}, "", "", apperror.Unauthorized("Sessiya yaroqsiz")
	}
	if item.Status != StatusActive {
		return Public{}, "", "", apperror.Forbidden("Hisob faol emas")
	}
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
	if use, _ := claims["use"].(string); use != "customer" {
		return uuid.Nil, apperror.Unauthorized("Token yaroqsiz")
	}
	sub, _ := claims["sub"].(string)
	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, apperror.Unauthorized("Token yaroqsiz")
	}
	return id, nil
}

func (s *AuthService) Me(ctx context.Context, id uuid.UUID) (Public, error) {
	item, err := s.Actor(ctx, id)
	if err != nil {
		return Public{}, err
	}
	return item.Public(), nil
}

func (s *AuthService) Actor(ctx context.Context, id uuid.UUID) (Customer, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Customer{}, apperror.Unauthorized("Sessiya yaroqsiz")
	}
	if item.Status != StatusActive {
		return Customer{}, apperror.Forbidden("Hisob faol emas")
	}
	return item, nil
}

func (s *AuthService) requireVerified(ctx context.Context, phone, purpose string) error {
	id, _, _, expiresAt, _, verifiedAt, usedAt, err := s.repo.LatestCode(ctx, phone, purpose)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.BadRequest("Avval SMS kodni tasdiqlang")
		}
		return apperror.Internal()
	}
	_ = id
	if usedAt != nil {
		return apperror.BadRequest("Kod allaqachon ishlatilgan")
	}
	if verifiedAt == nil {
		return apperror.BadRequest("Avval SMS kodni tasdiqlang")
	}
	if time.Now().After(expiresAt) {
		return apperror.BadRequest("Kod muddati tugagan")
	}
	return nil
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
		return apperror.New(http.StatusTooManyRequests, "Soatiga 5 tadan ko‘p SMS yuborib bo‘lmaydi")
	}
	st.hour = append(fresh, now)
	st.last = now
	return nil
}

func (s *AuthService) issue(ctx context.Context, item Customer) (string, string, error) {
	now := time.Now()
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  item.ID.String(),
		"role": "customer",
		"use":  "customer",
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

func NormalizePhone(phone string) (string, error) {
	p := strings.TrimSpace(phone)
	p = strings.ReplaceAll(p, " ", "")
	p = strings.ReplaceAll(p, "-", "")
	if strings.HasPrefix(p, "998") && len(p) == 12 {
		p = "+" + p
	}
	if !strings.HasPrefix(p, "+") && len(p) == 9 {
		p = "+998" + p
	}
	if !phoneRE.MatchString(p) {
		return "", apperror.BadRequest("Telefon formati: +998XXXXXXXXX")
	}
	return p, nil
}

func normalizePurpose(purpose string) (string, error) {
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		purpose = PurposeSetup
	}
	switch purpose {
	case PurposeRegister, PurposeSetup, PurposeReset, PurposeLogin:
		return purpose, nil
	default:
		return "", apperror.BadRequest("SMS maqsadi noto‘g‘ri")
	}
}

func simplePassword(password string) error {
	password = strings.TrimSpace(password)
	if len(password) < 6 {
		return apperror.BadRequest("Parol kamida 6 belgi bo‘lishi kerak")
	}
	if len(password) > 72 {
		return apperror.BadRequest("Parol juda uzun")
	}
	return nil
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
