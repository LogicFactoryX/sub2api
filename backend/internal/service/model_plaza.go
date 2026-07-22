package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const modelPlazaPriceDivisor = 5.0

var (
	ErrModelPlazaEntryNotFound = infraerrors.NotFound("MODEL_PLAZA_ENTRY_NOT_FOUND", "model plaza entry not found")
	ErrModelPlazaEntryExists   = infraerrors.Conflict("MODEL_PLAZA_ENTRY_EXISTS", "this model and group are already published")
)

type ModelPlazaPricing struct {
	InputPerMillion      *float64 `json:"input_per_million"`
	OutputPerMillion     *float64 `json:"output_per_million"`
	CacheWritePerMillion *float64 `json:"cache_write_per_million"`
	CacheReadPerMillion  *float64 `json:"cache_read_per_million"`
}

type ModelPlazaEntry struct {
	ID          int64    `json:"id"`
	ModelName   string   `json:"model_name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	GroupID     int64    `json:"group_id"`
	Tags        []string `json:"tags"`
	SortOrder   int      `json:"sort_order"`
	Enabled     bool     `json:"enabled"`

	GroupName        string            `json:"group_name"`
	Platform         string            `json:"platform"`
	GroupStatus      string            `json:"group_status"`
	RateMultiplier   float64           `json:"rate_multiplier"`
	IsExclusive      bool              `json:"is_exclusive"`
	IsMemberGroup    bool              `json:"is_member_group"`
	PricingAvailable bool              `json:"pricing_available"`
	OriginalPricing  ModelPlazaPricing `json:"original_pricing"`
	DisplayPricing   ModelPlazaPricing `json:"display_pricing"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ModelPlazaEntryInput struct {
	ModelName   string
	DisplayName string
	Description string
	GroupID     int64
	Tags        []string
	SortOrder   int
	Enabled     bool
}

type ModelPlazaRepository interface {
	List(ctx context.Context, publicOnly bool) ([]ModelPlazaEntry, error)
	GetByID(ctx context.Context, id int64) (*ModelPlazaEntry, error)
	Create(ctx context.Context, entry *ModelPlazaEntry) error
	Update(ctx context.Context, entry *ModelPlazaEntry) error
	Delete(ctx context.Context, id int64) error
}

type ModelPlazaService struct {
	repo      ModelPlazaRepository
	groupRepo GroupRepository
	billing   *BillingService
}

func NewModelPlazaService(repo ModelPlazaRepository, groupRepo GroupRepository, billing *BillingService) *ModelPlazaService {
	return &ModelPlazaService{repo: repo, groupRepo: groupRepo, billing: billing}
}

func (s *ModelPlazaService) ListPublic(ctx context.Context) ([]ModelPlazaEntry, error) {
	entries, err := s.repo.List(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("list public model plaza entries: %w", err)
	}
	s.enrichPricing(entries)
	return entries, nil
}

func (s *ModelPlazaService) ListAdmin(ctx context.Context) ([]ModelPlazaEntry, error) {
	entries, err := s.repo.List(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("list model plaza entries: %w", err)
	}
	s.enrichPricing(entries)
	return entries, nil
}

func (s *ModelPlazaService) Create(ctx context.Context, input ModelPlazaEntryInput) (*ModelPlazaEntry, error) {
	entry, err := s.prepareEntry(ctx, 0, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("create model plaza entry: %w", err)
	}
	s.enrichPricing([]ModelPlazaEntry{*entry})
	loaded, err := s.repo.GetByID(ctx, entry.ID)
	if err != nil {
		return nil, fmt.Errorf("load model plaza entry: %w", err)
	}
	s.enrichPricing([]ModelPlazaEntry{*loaded})
	s.enrichEntryPricing(loaded)
	return loaded, nil
}

func (s *ModelPlazaService) Update(ctx context.Context, id int64, input ModelPlazaEntryInput) (*ModelPlazaEntry, error) {
	if id <= 0 {
		return nil, ErrModelPlazaEntryNotFound
	}
	entry, err := s.prepareEntry(ctx, id, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("update model plaza entry: %w", err)
	}
	loaded, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load model plaza entry: %w", err)
	}
	s.enrichEntryPricing(loaded)
	return loaded, nil
}

func (s *ModelPlazaService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrModelPlazaEntryNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete model plaza entry: %w", err)
	}
	return nil
}

func (s *ModelPlazaService) prepareEntry(ctx context.Context, id int64, input ModelPlazaEntryInput) (*ModelPlazaEntry, error) {
	input.ModelName = strings.TrimSpace(input.ModelName)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Description = strings.TrimSpace(input.Description)
	if input.ModelName == "" || len(input.ModelName) > 255 {
		return nil, infraerrors.BadRequest("INVALID_MODEL_NAME", "model_name must contain 1 to 255 characters")
	}
	if len(input.DisplayName) > 255 {
		return nil, infraerrors.BadRequest("INVALID_DISPLAY_NAME", "display_name must contain at most 255 characters")
	}
	if len(input.Description) > 2000 {
		return nil, infraerrors.BadRequest("INVALID_DESCRIPTION", "description must contain at most 2000 characters")
	}
	if input.GroupID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_GROUP", "group_id is required")
	}
	group, err := s.groupRepo.GetByIDLite(ctx, input.GroupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("get model plaza group: %w", err)
	}
	if group.Status != StatusActive {
		return nil, infraerrors.BadRequest("INACTIVE_GROUP", "only active groups can be published")
	}

	return &ModelPlazaEntry{
		ID:          id,
		ModelName:   input.ModelName,
		DisplayName: input.DisplayName,
		Description: input.Description,
		GroupID:     input.GroupID,
		Tags:        normalizeModelPlazaTags(input.Tags),
		SortOrder:   input.SortOrder,
		Enabled:     input.Enabled,
	}, nil
}

func normalizeModelPlazaTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || len(tag) > 32 {
			continue
		}
		key := strings.ToLower(tag)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, tag)
		if len(out) == 8 {
			break
		}
	}
	return out
}

func (s *ModelPlazaService) enrichPricing(entries []ModelPlazaEntry) {
	for i := range entries {
		s.enrichEntryPricing(&entries[i])
	}
}

func (s *ModelPlazaService) enrichEntryPricing(entry *ModelPlazaEntry) {
	if entry == nil || s.billing == nil {
		return
	}
	pricing, err := s.billing.GetModelPricing(entry.ModelName)
	if err != nil || pricing == nil {
		return
	}

	factor := entry.RateMultiplier / modelPlazaPriceDivisor
	entry.PricingAvailable = true
	entry.OriginalPricing = modelPlazaPricingFromModel(pricing, 1)
	entry.DisplayPricing = modelPlazaPricingFromModel(pricing, factor)
}

func modelPlazaPricingFromModel(pricing *ModelPricing, factor float64) ModelPlazaPricing {
	perMillion := func(value float64, includeZero bool) *float64 {
		if value == 0 && !includeZero {
			return nil
		}
		converted := value * 1_000_000 * factor
		return &converted
	}
	return ModelPlazaPricing{
		InputPerMillion:      perMillion(pricing.InputPricePerToken, true),
		OutputPerMillion:     perMillion(pricing.OutputPricePerToken, true),
		CacheWritePerMillion: perMillion(pricing.CacheCreationPricePerToken, false),
		CacheReadPerMillion:  perMillion(pricing.CacheReadPricePerToken, false),
	}
}
