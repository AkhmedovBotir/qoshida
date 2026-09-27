package eskiz

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"qoshida/backend/internal/config"
)

const (
	defaultBaseURL     = "https://notify.eskiz.uz/api"
	defaultSender      = "4546"
	tokenRefreshPeriod = 29 * 24 * time.Hour
)

var (
	ErrConfigInvalid   = errors.New("eskiz sozlamalari to'liq emas")
	ErrAuthFailed      = errors.New("SMS xizmatiga ulanishda xatolik yuz berdi")
	ErrSendFailed      = errors.New("SMS yuborishda xatolik yuz berdi")
	ErrUnknownTemplate = errors.New("noma'lum SMS shablon")
	ErrPhoneInvalid    = errors.New("telefon raqami noto'g'ri")
	ErrCodeRequired    = errors.New("tasdiqlash kodi majburiy")
)

var phoneCleaner = regexp.MustCompile(`[+\s\-()]`)

type Service struct {
	email    string
	password string
	baseURL  string
	sender   string
	client   *http.Client

	mu             sync.Mutex
	token          string
	tokenExpiresAt time.Time
}

type SendResult struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Status    string `json:"status"`
}

func New(cfg *config.Config) *Service {
	baseURL := strings.TrimSpace(cfg.EskizBaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	sender := strings.TrimSpace(cfg.EskizSender)
	if sender == "" {
		sender = defaultSender
	}
	timeout := cfg.EskizTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Service{
		email:    strings.TrimSpace(cfg.EskizEmail),
		password: strings.TrimSpace(cfg.EskizPassword),
		baseURL:  strings.TrimRight(baseURL, "/"),
		sender:   sender,
		client:   &http.Client{Timeout: timeout},
	}
}

func (s *Service) Enabled() bool {
	return s != nil && s.email != "" && s.password != ""
}

func (s *Service) ValidateConfig() error {
	if !s.Enabled() {
		return ErrConfigInvalid
	}
	return nil
}

func (s *Service) SendCode(ctx context.Context, phone string, tpl Template, code string) (*SendResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrCodeRequired
	}
	message, err := Render(tpl, code)
	if err != nil {
		return nil, err
	}
	return s.SendSMS(ctx, phone, message)
}

func (s *Service) SendMarketplaceRegister(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateMarketplaceRegister, code)
}

func (s *Service) SendMarketplaceLogin(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateMarketplaceLogin, code)
}

func (s *Service) SendMarketplaceReset(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateMarketplaceReset, code)
}

func (s *Service) SendDeviceVerify(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateDeviceVerify, code)
}

func (s *Service) SendContragentPassword(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateContragentPassword, code)
}

func (s *Service) SendPunktPassword(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplatePunktPassword, code)
}

func (s *Service) SendAgentPassword(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateAgentPassword, code)
}

func (s *Service) SendDeliveryPassword(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateDeliveryPassword, code)
}

func (s *Service) SendManagerPassword(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateManagerPassword, code)
}

func (s *Service) SendLocalShopPassword(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateLocalShopPassword, code)
}

func (s *Service) SendShopDirectorPassword(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateShopDirectorPassword, code)
}

func (s *Service) SendSellerPassword(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateSellerPassword, code)
}

func (s *Service) SendServicePassword(ctx context.Context, phone, code string) (*SendResult, error) {
	return s.SendCode(ctx, phone, TemplateServicePassword, code)
}

func (s *Service) SendSMS(ctx context.Context, phone, message string) (*SendResult, error) {
	return s.sendSMS(ctx, phone, message, true)
}

func (s *Service) sendSMS(ctx context.Context, phone, message string, allowRetry bool) (*SendResult, error) {
	if err := s.ValidateConfig(); err != nil {
		return nil, err
	}
	finalPhone, err := NormalizePhone(phone)
	if err != nil {
		return nil, err
	}
	token, err := s.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	body, contentType, err := formBody(map[string]string{
		"mobile_phone": finalPhone,
		"message":      message,
		"from":         s.sender,
	})
	if err != nil {
		return nil, ErrSendFailed
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/message/sms/send", body)
	if err != nil {
		return nil, ErrSendFailed
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, ErrSendFailed
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Status  string `json:"status"`
		ID      any    `json:"id"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(raw, &parsed)

	if resp.StatusCode == http.StatusUnauthorized && allowRetry {
		s.clearToken()
		if _, refreshErr := s.refreshToken(ctx); refreshErr != nil {
			if _, loginErr := s.login(ctx); loginErr != nil {
				return nil, loginErr
			}
		}
		return s.sendSMS(ctx, phone, message, false)
	}

	if parsed.Status == "success" || parsed.Status == "waiting" || parsed.ID != nil ||
		strings.Contains(strings.ToLower(parsed.Message), "waiting") {
		status := parsed.Status
		if status == "" {
			status = "sent"
		}
		return &SendResult{Success: true, MessageID: anyString(parsed.ID), Status: status}, nil
	}
	if parsed.Message != "" {
		return nil, fmt.Errorf("%w: %s", ErrSendFailed, parsed.Message)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("%w: %s", ErrSendFailed, strings.TrimSpace(string(raw)))
	}
	return &SendResult{Success: true, Status: "sent"}, nil
}

func (s *Service) GetToken(ctx context.Context) (string, error) {
	if err := s.ValidateConfig(); err != nil {
		return "", err
	}

	s.mu.Lock()
	if s.token != "" && time.Now().Before(s.tokenExpiresAt) {
		token := s.token
		s.mu.Unlock()
		return token, nil
	}
	s.mu.Unlock()
	return s.login(ctx)
}

func (s *Service) login(ctx context.Context) (string, error) {
	body, contentType, err := formBody(map[string]string{
		"email":    s.email,
		"password": s.password,
	})
	if err != nil {
		return "", ErrAuthFailed
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/auth/login", body)
	if err != nil {
		return "", ErrAuthFailed
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", ErrAuthFailed
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(raw, &parsed)
	if resp.StatusCode >= http.StatusBadRequest || parsed.Data.Token == "" {
		return "", fmt.Errorf("%w: %s", ErrAuthFailed, strings.TrimSpace(string(raw)))
	}

	s.storeToken(parsed.Data.Token)
	return parsed.Data.Token, nil
}

func (s *Service) refreshToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	current := s.token
	s.mu.Unlock()
	if current == "" {
		return "", ErrAuthFailed
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, s.baseURL+"/auth/refresh", nil)
	if err != nil {
		return "", ErrAuthFailed
	}
	req.Header.Set("Authorization", "Bearer "+current)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", ErrAuthFailed
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
		Token string `json:"token"`
	}
	_ = json.Unmarshal(raw, &parsed)
	token := parsed.Data.Token
	if token == "" {
		token = parsed.Token
	}
	if resp.StatusCode >= http.StatusBadRequest || token == "" {
		return "", ErrAuthFailed
	}
	s.storeToken(token)
	return token, nil
}

func (s *Service) NewCode() string {
	var b [4]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		n := time.Now().UnixNano() % 1_000_000
		if n < 0 {
			n = -n
		}
		return fmt.Sprintf("%06d", n)
	}
	n := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if n < 0 {
		n = -n
	}
	return fmt.Sprintf("%06d", n%1_000_000)
}

func (s *Service) storeToken(token string) {
	s.mu.Lock()
	s.token = token
	s.tokenExpiresAt = time.Now().Add(tokenRefreshPeriod)
	s.mu.Unlock()
}

func (s *Service) clearToken() {
	s.mu.Lock()
	s.token = ""
	s.tokenExpiresAt = time.Time{}
	s.mu.Unlock()
}

func NormalizePhone(phone string) (string, error) {
	formatted := phoneCleaner.ReplaceAllString(strings.TrimSpace(phone), "")
	if formatted == "" {
		return "", ErrPhoneInvalid
	}
	if strings.HasPrefix(formatted, "998") {
		if len(formatted) != 12 {
			return "", ErrPhoneInvalid
		}
		return formatted, nil
	}
	if len(formatted) == 9 {
		return "998" + formatted, nil
	}
	return "", ErrPhoneInvalid
}

func anyString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func formBody(fields map[string]string) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := w.WriteField(key, value); err != nil {
			_ = w.Close()
			return nil, "", err
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &buf, w.FormDataContentType(), nil
}
