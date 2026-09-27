package admin

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"qoshida/backend/internal/shared/apperror"
)

var (
	phoneRE    = regexp.MustCompile(`^\+998[0-9]{9}$`)
	usernameRE = regexp.MustCompile(`^[a-zA-Z0-9._]{3,32}$`)
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal()
	}
	out := make([]Public, 0, len(items))
	for _, item := range items {
		out = append(out, item.Public())
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Public, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Admin topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	return item.Public(), nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Public, error) {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Username = strings.TrimSpace(req.Username)

	if err := validateIdentity(req.FirstName, req.LastName, req.Phone, req.Username); err != nil {
		return Public{}, err
	}
	if err := validatePassword(req.Password); err != nil {
		return Public{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return Public{}, apperror.Internal()
	}

	item, err := s.repo.Create(ctx, Admin{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Username:     req.Username,
		Role:         RoleAdmin,
		PasswordHash: string(hash),
	})
	if err != nil {
		return Public{}, mapWriteError(err)
	}
	return item.Public(), nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Admin topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if current.Role == RoleGeneral {
		return Public{}, apperror.Forbidden("General adminni o'zgartirish mumkin emas")
	}

	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Username = strings.TrimSpace(req.Username)

	if err = validateIdentity(req.FirstName, req.LastName, req.Phone, req.Username); err != nil {
		return Public{}, err
	}

	hash := current.PasswordHash
	if strings.TrimSpace(req.Password) != "" {
		if err = validatePassword(req.Password); err != nil {
			return Public{}, err
		}
		raw, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			return Public{}, apperror.Internal()
		}
		hash = string(raw)
	}

	item, err := s.repo.Update(ctx, Admin{
		ID:           id,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Username:     req.Username,
		PasswordHash: hash,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.Forbidden("General adminni o'zgartirish mumkin emas")
		}
		return Public{}, mapWriteError(err)
	}
	return item.Public(), nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	if id == actorID {
		return apperror.Forbidden("O'zingizni o'chira olmaysiz")
	}
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Admin topilmadi")
		}
		return apperror.Internal()
	}
	if current.Role == RoleGeneral {
		return apperror.Forbidden("General adminni o'chirish mumkin emas")
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.Forbidden("General adminni o'chirish mumkin emas")
		}
		return apperror.Internal()
	}
	return nil
}

func (s *Service) CreateGeneral(ctx context.Context, req CreateRequest) (Public, error) {
	exists, err := s.repo.HasGeneral(ctx)
	if err != nil {
		return Public{}, err
	}
	if exists {
		return Public{}, apperror.Conflict("General admin allaqachon mavjud")
	}

	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Username = strings.TrimSpace(req.Username)

	if err = validateIdentity(req.FirstName, req.LastName, req.Phone, req.Username); err != nil {
		return Public{}, err
	}
	if err = validatePassword(req.Password); err != nil {
		return Public{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return Public{}, err
	}

	item, err := s.repo.Create(ctx, Admin{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Username:     req.Username,
		Role:         RoleGeneral,
		PasswordHash: string(hash),
	})
	if err != nil {
		return Public{}, err
	}
	return item.Public(), nil
}

func (s *Service) Stats(ctx context.Context) (map[string]int, error) {
	n, err := s.repo.Count(ctx)
	if err != nil {
		return nil, apperror.Internal()
	}
	return map[string]int{"admins": n}, nil
}

func validateIdentity(firstName, lastName, phone, username string) error {
	if len(firstName) < 2 || len(firstName) > 80 {
		return apperror.BadRequest("Ism 2-80 belgi oralig'ida bo'lishi kerak")
	}
	if len(lastName) < 2 || len(lastName) > 80 {
		return apperror.BadRequest("Familiya 2-80 belgi oralig'ida bo'lishi kerak")
	}
	if !phoneRE.MatchString(phone) {
		return apperror.BadRequest("Telefon formati: +998XXXXXXXXX")
	}
	if !usernameRE.MatchString(username) {
		return apperror.BadRequest("Username 3-32 belgi: harf, raqam, nuqta yoki pastki chiziq")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 10 || len(password) > 72 {
		return apperror.BadRequest("Parol 10-72 belgi oralig'ida bo'lishi kerak")
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return apperror.BadRequest("Parolda kamida bitta harf va bitta raqam bo'lishi kerak")
	}
	return nil
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.Conflict("Username yoki telefon allaqachon mavjud")
	}
	return apperror.Internal()
}
