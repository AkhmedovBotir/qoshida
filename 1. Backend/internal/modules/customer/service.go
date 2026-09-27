package customer

import (
	"context"
	"database/sql"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"qoshida/backend/internal/modules/region"
	"qoshida/backend/internal/shared/apperror"
)

type Service struct {
	repo    *Repository
	regions *region.Repository
}

func NewService(repo *Repository, regions *region.Repository) *Service {
	return &Service{repo: repo, regions: regions}
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, req ProfileRequest) (Public, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Public{}, apperror.NotFound("Mijoz topilmadi")
	}
	fields, err := s.ParseProfile(ctx, req, true)
	if err != nil {
		return Public{}, err
	}
	item.FirstName = fields.FirstName
	item.LastName = fields.LastName
	item.Name = fields.Name
	item.BirthDate = fields.BirthDate
	item.RegionID = fields.RegionID
	item.DistrictID = fields.DistrictID
	item.MFYID = fields.MFYID
	item.Address = fields.Address
	if req.ClearAvatar {
		removeAvatarFile(item.Avatar)
		item.Avatar = ""
	} else {
		avatar, err := saveAvatar(item.ID, req.Avatar, item.Avatar)
		if err != nil {
			return Public{}, err
		}
		item.Avatar = avatar
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	return updated.Public(), nil
}

func (s *Service) ParseProfile(ctx context.Context, req ProfileRequest, required bool) (ProfileFields, error) {
	first := strings.Join(strings.Fields(strings.TrimSpace(req.FirstName)), " ")
	last := strings.Join(strings.Fields(strings.TrimSpace(req.LastName)), " ")
	if first == "" && last == "" {
		parts := strings.Fields(strings.TrimSpace(req.Name))
		if len(parts) > 0 {
			first = parts[0]
		}
		if len(parts) > 1 {
			last = strings.Join(parts[1:], " ")
		}
	}
	if n := utf8.RuneCountInString(first); n < 2 || n > 60 {
		return ProfileFields{}, apperror.BadRequest("Ism 2-60 belgi oralig‘ida bo‘lishi kerak")
	}
	if n := utf8.RuneCountInString(last); n < 2 || n > 60 {
		return ProfileFields{}, apperror.BadRequest("Familiya 2-60 belgi oralig‘ida bo‘lishi kerak")
	}
	birth, err := parseBirthDate(req.BirthDate, required)
	if err != nil {
		return ProfileFields{}, err
	}
	regionID, districtID, mfyID, err := s.parseGeo(ctx, req.RegionID, req.DistrictID, req.MFYID, required)
	if err != nil {
		return ProfileFields{}, err
	}
	name := strings.TrimSpace(first + " " + last)
	address := strings.TrimSpace(req.Address)
	if utf8.RuneCountInString(address) > 300 {
		return ProfileFields{}, apperror.BadRequest("Manzil 300 belgidan oshmasin")
	}
	return ProfileFields{
		FirstName:  first,
		LastName:   last,
		Name:       name,
		BirthDate:  birth,
		RegionID:   regionID,
		DistrictID: districtID,
		MFYID:      mfyID,
		Address:    address,
	}, nil
}

func (s *Service) parseGeo(ctx context.Context, regionID, districtID, mfyID string, required bool) (*uuid.UUID, *uuid.UUID, *uuid.UUID, error) {
	rid := parseOptionalUUID(regionID)
	did := parseOptionalUUID(districtID)
	mid := parseOptionalUUID(mfyID)
	if required && (rid == nil || did == nil || mid == nil) {
		return nil, nil, nil, apperror.BadRequest("Viloyat, tuman va MFY tanlanishi shart")
	}
	if rid != nil {
		reg, err := s.regions.FindByID(ctx, *rid)
		if err != nil || reg.Type != region.TypeRegion {
			return nil, nil, nil, apperror.BadRequest("Viloyat noto‘g‘ri")
		}
	}
	if did != nil {
		dis, err := s.regions.FindByID(ctx, *did)
		if err != nil || dis.Type != region.TypeDistrict {
			return nil, nil, nil, apperror.BadRequest("Tuman noto‘g‘ri")
		}
		if rid != nil && (dis.ParentID == nil || *dis.ParentID != *rid) {
			return nil, nil, nil, apperror.BadRequest("Tuman tanlangan viloyatga tegishli emas")
		}
	}
	if mid != nil {
		mfy, err := s.regions.FindByID(ctx, *mid)
		if err != nil || mfy.Type != region.TypeMFY {
			return nil, nil, nil, apperror.BadRequest("MFY noto‘g‘ri")
		}
		if did != nil && (mfy.ParentID == nil || *mfy.ParentID != *did) {
			return nil, nil, nil, apperror.BadRequest("MFY tanlangan tumanga tegishli emas")
		}
	}
	return rid, did, mid, nil
}

func parseBirthDate(raw string, required bool) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return nil, apperror.BadRequest("Tug‘ilgan sana kiritilishi shart")
		}
		return nil, nil
	}
	day, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, apperror.BadRequest("Tug‘ilgan sana formati YYYY-MM-DD")
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if day.After(today) {
		return nil, apperror.BadRequest("Tug‘ilgan sana kelajakda bo‘lishi mumkin emas")
	}
	age := today.Year() - day.Year()
	if today.YearDay() < day.YearDay() {
		age--
	}
	if age < 14 {
		return nil, apperror.BadRequest("Yosh 14 dan kichik bo‘lmasligi kerak")
	}
	if age > 100 {
		return nil, apperror.BadRequest("Tug‘ilgan sana noto‘g‘ri")
	}
	return &day, nil
}

func parseOptionalUUID(raw string) *uuid.UUID {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func nullTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func scanBirth(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}
