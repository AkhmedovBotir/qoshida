package serviceprovider

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
)

const (
	IdentNone     = "none"
	IdentPending  = "pending"
	IdentApproved = "approved"
	IdentRejected = "rejected"

	identEntity = "service_provider_identification"

	maxIdentDocs     = 5
	maxAboutRunes    = 1000
	maxAddressRunes  = 300
	minAddressRunes  = 10
	maxFullNameRunes = 160
	minFullNameRunes = 5
)

var (
	pinflRE    = regexp.MustCompile(`^[0-9]{14}$`)
	passportRE = regexp.MustCompile(`^[A-Z]{2}[0-9]{7}$`)
)

type IdentDocument struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Path     string `json:"path"`
	FileName string `json:"file_name"`
}

type Identification struct {
	ID               uuid.UUID
	ProviderID       uuid.UUID
	FullName         string
	PINFL            string
	Passport         string
	BirthDate        time.Time
	Address          string
	ExperienceYears  int
	About            string
	PassportImage    string
	SelfieImage      string
	Documents        []IdentDocument
	Status           string
	RejectionNote    string
	ReviewedByName   string
	ReviewedByRole   string
	ReviewedAt       *time.Time
	SubmittedAt      time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ProviderName     string
	Phone            string
	ActivityName     string
	RegionName       string
	DistrictName     string
	MFYName          string
	RegionID         uuid.UUID
	DistrictID       uuid.UUID
	MFYID            uuid.UUID
}

type IdentPublic struct {
	ID               string          `json:"id,omitempty"`
	ProviderID       string          `json:"provider_id"`
	FullName         string          `json:"full_name,omitempty"`
	PINFL            string          `json:"pinfl,omitempty"`
	Passport         string          `json:"passport,omitempty"`
	BirthDate        string          `json:"birth_date,omitempty"`
	Address          string          `json:"address,omitempty"`
	ExperienceYears  int             `json:"experience_years,omitempty"`
	About            string          `json:"about,omitempty"`
	PassportImage    string          `json:"passport_image,omitempty"`
	SelfieImage      string          `json:"selfie_image,omitempty"`
	Documents        []IdentDocument `json:"documents,omitempty"`
	Status           string          `json:"status"`
	RejectionNote    string          `json:"rejection_note,omitempty"`
	ReviewedByName   string          `json:"reviewed_by_name,omitempty"`
	ReviewedByRole   string          `json:"reviewed_by_role,omitempty"`
	ReviewedAt       string          `json:"reviewed_at,omitempty"`
	SubmittedAt      string          `json:"submitted_at,omitempty"`
	ProviderName     string          `json:"provider_name,omitempty"`
	Phone            string          `json:"phone,omitempty"`
	ActivityName     string          `json:"activity_name,omitempty"`
	RegionName       string          `json:"region_name,omitempty"`
	DistrictName     string          `json:"district_name,omitempty"`
	MFYName          string          `json:"mfy_name,omitempty"`
	CreatedAt        string          `json:"created_at,omitempty"`
	Audit            *audit.Trail    `json:"audit,omitempty"`
}

type IdentDocumentInput struct {
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	FileName string `json:"file_name"`
	Path     string `json:"path"`
	Content  string `json:"content"`
}

type IdentSubmitRequest struct {
	FullName        string               `json:"full_name"`
	PINFL           string               `json:"pinfl"`
	Passport        string               `json:"passport"`
	BirthDate       string               `json:"birth_date"`
	Address         string               `json:"address"`
	ExperienceYears int                  `json:"experience_years"`
	About           string               `json:"about"`
	PassportImage   string               `json:"passport_image"`
	SelfieImage     string               `json:"selfie_image"`
	Documents       []IdentDocumentInput `json:"documents"`
}

type IdentListQuery struct {
	Status     string
	Query      string
	Page       int
	Limit      int
	RegionID   *uuid.UUID
	DistrictID *uuid.UUID
	MFYID      *uuid.UUID
}

type IdentRejectRequest struct {
	Note string `json:"note"`
}

type IdentListResult struct {
	Items []IdentPublic `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

func (item Identification) Public() IdentPublic {
	out := IdentPublic{
		ID:              item.ID.String(),
		ProviderID:      item.ProviderID.String(),
		FullName:        item.FullName,
		PINFL:           item.PINFL,
		Passport:        item.Passport,
		BirthDate:       item.BirthDate.Format("2006-01-02"),
		Address:         item.Address,
		ExperienceYears: item.ExperienceYears,
		About:           item.About,
		PassportImage:   item.PassportImage,
		SelfieImage:     item.SelfieImage,
		Documents:       item.Documents,
		Status:          item.Status,
		RejectionNote:   item.RejectionNote,
		ReviewedByName:  item.ReviewedByName,
		ReviewedByRole:  item.ReviewedByRole,
		ProviderName:    item.ProviderName,
		Phone:           item.Phone,
		ActivityName:    item.ActivityName,
		RegionName:      item.RegionName,
		DistrictName:    item.DistrictName,
		MFYName:         item.MFYName,
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
	}
	if item.Documents == nil {
		out.Documents = []IdentDocument{}
	}
	if item.ReviewedAt != nil {
		out.ReviewedAt = item.ReviewedAt.UTC().Format(time.RFC3339)
	}
	if !item.SubmittedAt.IsZero() {
		out.SubmittedAt = item.SubmittedAt.UTC().Format(time.RFC3339)
	}
	return out
}

func emptyIdent(providerID uuid.UUID) IdentPublic {
	return IdentPublic{ProviderID: providerID.String(), Status: IdentNone, Documents: []IdentDocument{}}
}

func (s *Service) GetOwnIdentification(ctx context.Context, providerID uuid.UUID) (IdentPublic, error) {
	item, err := s.repo.FindIdentByProvider(ctx, providerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return emptyIdent(providerID), nil
		}
		return IdentPublic{}, apperror.Internal()
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, identEntity, item.ID)
	attachIdentReview(&pub)
	return pub, nil
}

func (s *Service) SubmitIdentification(ctx context.Context, providerID uuid.UUID, req IdentSubmitRequest) (IdentPublic, error) {
	current, err := s.repo.FindIdentByProvider(ctx, providerID)
	exists := err == nil
	if err != nil && !errors.Is(err, ErrNotFound) {
		return IdentPublic{}, apperror.Internal()
	}
	if exists && current.Status == IdentPending {
		return IdentPublic{}, apperror.BadRequest("Identifikatsiya ko‘rib chiqilmoqda. Javobni kuting")
	}
	if exists && current.Status == IdentApproved {
		return IdentPublic{}, apperror.Forbidden("Identifikatsiya allaqachon tasdiqlangan")
	}

	built, err := s.buildIdentification(providerID, req, current, exists)
	if err != nil {
		return IdentPublic{}, err
	}

	var saved Identification
	if exists {
		built.ID = current.ID
		built.CreatedAt = current.CreatedAt
		saved, err = s.repo.UpdateIdent(ctx, built)
	} else {
		built.ID = uuid.New()
		saved, err = s.repo.CreateIdent(ctx, built)
	}
	if err != nil {
		return IdentPublic{}, mapIdentWriteError(err)
	}
	if exists {
		cleanupIdentFiles(current, saved)
	}

	pub := saved.Public()
	if exists {
		pub.Audit = audit.AfterUpdate(ctx, identEntity, saved.ID)
	} else {
		pub.Audit = audit.AfterCreate(ctx, identEntity, saved.ID)
	}
	return pub, nil
}

func (s *Service) ListIdentifications(ctx context.Context, q IdentListQuery, scope *Scope) (IdentListResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}
	q.Query = strings.TrimSpace(q.Query)
	q.Status = strings.TrimSpace(q.Status)
	applyIdentListScope(&q, scope)
	items, total, err := s.repo.ListIdent(ctx, q)
	if err != nil {
		return IdentListResult{}, apperror.Internal()
	}
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	trails := audit.BindMany(ctx, identEntity, ids)
	out := make([]IdentPublic, 0, len(items))
	for _, item := range items {
		pub := item.Public()
		if t, ok := trails[item.ID]; ok {
			pub.Audit = &t
		}
		attachIdentReview(&pub)
		out = append(out, pub)
	}
	return IdentListResult{Items: out, Total: total, Page: q.Page, Limit: q.Limit}, nil
}

func (s *Service) GetIdentification(ctx context.Context, id uuid.UUID, scope *Scope) (IdentPublic, error) {
	item, err := s.repo.FindIdentByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return IdentPublic{}, apperror.NotFound("Identifikatsiya topilmadi")
		}
		return IdentPublic{}, apperror.Internal()
	}
	if err = assertIdentInScope(item, scope); err != nil {
		return IdentPublic{}, err
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, identEntity, item.ID)
	attachIdentReview(&pub)
	return pub, nil
}

func (s *Service) ApproveIdentification(ctx context.Context, id uuid.UUID, scope *Scope) (IdentPublic, error) {
	current, err := s.requirePendingIdent(ctx, id, scope)
	if err != nil {
		return IdentPublic{}, err
	}
	current.Status = IdentApproved
	current.RejectionNote = ""
	stampIdentReview(ctx, &current)
	updated, err := s.repo.UpdateIdentReview(ctx, current)
	if err != nil {
		return IdentPublic{}, apperror.Internal()
	}
	pub := updated.Public()
	pub.Audit = audit.AfterApprove(ctx, identEntity, updated.ID)
	attachIdentReview(&pub)
	return pub, nil
}

func (s *Service) RejectIdentification(ctx context.Context, id uuid.UUID, note string, scope *Scope) (IdentPublic, error) {
	note = strings.TrimSpace(note)
	if utf8.RuneCountInString(note) < 3 {
		return IdentPublic{}, apperror.BadRequest("Bekor qilish sababi kamida 3 belgi bo‘lishi kerak")
	}
	if utf8.RuneCountInString(note) > 500 {
		return IdentPublic{}, apperror.BadRequest("Bekor qilish sababi 500 belgidan oshmasin")
	}
	current, err := s.requirePendingIdent(ctx, id, scope)
	if err != nil {
		return IdentPublic{}, err
	}
	current.Status = IdentRejected
	current.RejectionNote = note
	stampIdentReview(ctx, &current)
	updated, err := s.repo.UpdateIdentReview(ctx, current)
	if err != nil {
		return IdentPublic{}, apperror.Internal()
	}
	pub := updated.Public()
	pub.Audit = audit.AfterReject(ctx, identEntity, updated.ID)
	attachIdentReview(&pub)
	return pub, nil
}

func (s *Service) requirePendingIdent(ctx context.Context, id uuid.UUID, scope *Scope) (Identification, error) {
	current, err := s.repo.FindIdentByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Identification{}, apperror.NotFound("Identifikatsiya topilmadi")
		}
		return Identification{}, apperror.Internal()
	}
	if err = assertIdentInScope(current, scope); err != nil {
		return Identification{}, err
	}
	if current.Status != IdentPending {
		return Identification{}, apperror.BadRequest("Faqat kutilayotgan identifikatsiyani tasdiqlash yoki bekor qilish mumkin")
	}
	return current, nil
}

func (s *Service) buildIdentification(providerID uuid.UUID, req IdentSubmitRequest, current Identification, exists bool) (Identification, error) {
	fullName := compactSpace(req.FullName)
	if n := utf8.RuneCountInString(fullName); n < minFullNameRunes || n > maxFullNameRunes {
		return Identification{}, apperror.BadRequest("F.I.Sh 5-160 belgi oralig‘ida bo‘lishi kerak")
	}
	pinfl := strings.TrimSpace(req.PINFL)
	if !pinflRE.MatchString(pinfl) {
		return Identification{}, apperror.BadRequest("JSHSHIR 14 ta raqam bo‘lishi kerak")
	}
	passport := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(req.Passport), " ", ""))
	if !passportRE.MatchString(passport) {
		return Identification{}, apperror.BadRequest("Pasport formati: AA1234567")
	}
	birth, err := time.Parse("2006-01-02", strings.TrimSpace(req.BirthDate))
	if err != nil {
		return Identification{}, apperror.BadRequest("Tug‘ilgan sana noto‘g‘ri")
	}
	if err = validateBirthDate(birth); err != nil {
		return Identification{}, err
	}
	address := compactSpace(req.Address)
	if n := utf8.RuneCountInString(address); n < minAddressRunes || n > maxAddressRunes {
		return Identification{}, apperror.BadRequest("Manzil 10-300 belgi oralig‘ida bo‘lishi kerak")
	}
	if req.ExperienceYears < 0 || req.ExperienceYears > 70 {
		return Identification{}, apperror.BadRequest("Ish tajribasi 0-70 yil oralig‘ida bo‘lishi kerak")
	}
	about := strings.TrimSpace(req.About)
	if utf8.RuneCountInString(about) > maxAboutRunes {
		return Identification{}, apperror.BadRequest("Qo‘shimcha ma’lumot 1000 belgidan oshmasin")
	}

	prevPassport := ""
	prevSelfie := ""
	if exists {
		prevPassport = current.PassportImage
		prevSelfie = current.SelfieImage
	}
	passportImage, err := saveIdentImage(providerID, "passport", req.PassportImage, prevPassport)
	if err != nil {
		return Identification{}, err
	}
	if passportImage == "" {
		return Identification{}, apperror.BadRequest("Pasport rasmi majburiy")
	}
	selfieImage, err := saveIdentImage(providerID, "selfie", req.SelfieImage, prevSelfie)
	if err != nil {
		return Identification{}, err
	}
	if selfieImage == "" {
		return Identification{}, apperror.BadRequest("Selfi rasmi majburiy")
	}

	docs, err := saveIdentDocuments(providerID, req.Documents, current.Documents)
	if err != nil {
		return Identification{}, err
	}

	now := time.Now().UTC()
	return Identification{
		ProviderID:      providerID,
		FullName:        fullName,
		PINFL:           pinfl,
		Passport:        passport,
		BirthDate:       birth,
		Address:         address,
		ExperienceYears: req.ExperienceYears,
		About:           about,
		PassportImage:   passportImage,
		SelfieImage:     selfieImage,
		Documents:       docs,
		Status:          IdentPending,
		RejectionNote:   "",
		SubmittedAt:     now,
	}, nil
}

func validateBirthDate(birth time.Time) error {
	today := time.Now().UTC()
	if birth.After(today) {
		return apperror.BadRequest("Tug‘ilgan sana kelajakda bo‘lishi mumkin emas")
	}
	age := today.Year() - birth.Year()
	if today.YearDay() < birth.YearDay() {
		age--
	}
	if age < 18 {
		return apperror.BadRequest("Yosh 18 dan kichik bo‘lmasligi kerak")
	}
	if age > 80 {
		return apperror.BadRequest("Tug‘ilgan sana noto‘g‘ri")
	}
	return nil
}

func applyIdentListScope(q *IdentListQuery, scope *Scope) {
	if scope == nil {
		return
	}
	q.RegionID = &scope.RegionID
	if scope.DistrictID != nil {
		q.DistrictID = scope.DistrictID
	}
	if scope.MFYID != nil {
		q.MFYID = scope.MFYID
	}
}

func assertIdentInScope(item Identification, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if item.RegionID != scope.RegionID {
		return apperror.Forbidden("Faqat o‘z viloyatingizdagi identifikatsiyani ko‘rish mumkin")
	}
	if scope.DistrictID != nil && item.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Faqat o‘z tumaningizdagi identifikatsiyani ko‘rish mumkin")
	}
	if scope.MFYID != nil && item.MFYID != *scope.MFYID {
		return apperror.Forbidden("Faqat o‘z MFYingizdagi identifikatsiyani ko‘rish mumkin")
	}
	return nil
}

func attachIdentReview(pub *IdentPublic) {
	if pub.ReviewedByName == "" {
		return
	}
	if pub.Audit == nil {
		pub.Audit = &audit.Trail{}
	}
	entry := &audit.Entry{Name: pub.ReviewedByName, Role: pub.ReviewedByRole, At: pub.ReviewedAt}
	if pub.Status == IdentRejected && pub.Audit.Rejected == nil {
		pub.Audit.Rejected = entry
	}
	if pub.Status == IdentApproved && pub.Audit.Approved == nil {
		pub.Audit.Approved = entry
	}
}

func stampIdentReview(ctx context.Context, item *Identification) {
	now := time.Now().UTC()
	item.ReviewedAt = &now
	if actor, ok := audit.From(ctx); ok {
		item.ReviewedByName = actor.Name
		item.ReviewedByRole = actor.Role
		return
	}
	item.ReviewedByName = "Tizim"
	item.ReviewedByRole = "Admin"
}

func compactSpace(raw string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
}

func marshalDocs(docs []IdentDocument) []byte {
	if docs == nil {
		docs = []IdentDocument{}
	}
	raw, err := json.Marshal(docs)
	if err != nil {
		return []byte("[]")
	}
	return raw
}

func unmarshalDocs(raw []byte) []IdentDocument {
	if len(raw) == 0 {
		return []IdentDocument{}
	}
	var docs []IdentDocument
	if err := json.Unmarshal(raw, &docs); err != nil || docs == nil {
		return []IdentDocument{}
	}
	return docs
}

func mapIdentWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.Conflict("Bu JSHSHIR allaqachon ro‘yxatdan o‘tgan")
	}
	return apperror.Internal()
}
