package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"CampusWatch/backend/internal/model"
)

// CampusRepository defines the persistence contract for campuses.
type CampusRepository interface {
	Create(ctx context.Context, campus model.Campus) (*model.Campus, error)
	List(ctx context.Context) ([]model.Campus, error)
	GetByID(ctx context.Context, id string) (*model.Campus, error)
	Update(ctx context.Context, id string, input model.UpdateCampusRequest) (*model.Campus, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// CampusService provides the business logic for campus operations.
type CampusService struct {
	repo CampusRepository
}

// NewCampusService creates a campus service with a repository implementation.
func NewCampusService(repo CampusRepository) *CampusService {
	return &CampusService{repo: repo}
}

// Create creates a new campus after validating key fields.
func (s *CampusService) Create(ctx context.Context, institutionID, name, slug, status string) (*model.Campus, error) {
	trimmedInstitutionID := strings.TrimSpace(institutionID)
	if trimmedInstitutionID == "" {
		return nil, errors.New("institution id is required")
	}

	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, errors.New("campus name is required")
	}

	normalizedSlug := campusSlugify(slug)
	if normalizedSlug == "" {
		normalizedSlug = campusSlugify(trimmedName)
	}
	if normalizedSlug == "" {
		return nil, errors.New("campus slug is required")
	}

	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus == "" {
		normalizedStatus = "active"
	}
	if !isCampusStatusValid(normalizedStatus) {
		return nil, fmt.Errorf("invalid campus status %q", normalizedStatus)
	}

	campus := model.Campus{
		InstitutionID: trimmedInstitutionID,
		Name:          trimmedName,
		Slug:          normalizedSlug,
		Status:        normalizedStatus,
	}

	return s.repo.Create(ctx, campus)
}

// List fetches all campuses records.
func (s *CampusService) List(ctx context.Context) ([]model.Campus, error) {
	return s.repo.List(ctx)
}

// GetByID loads one campus by ID.
func (s *CampusService) GetByID(ctx context.Context, id string) (*model.Campus, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("campus id is required")
	}
	return s.repo.GetByID(ctx, id)
}

// Update applies partial field updates to a campus.
func (s *CampusService) Update(ctx context.Context, id string, input model.UpdateCampusRequest) (*model.Campus, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("campus id is required")
	}

	if input.InstitutionID != nil {
		trimmed := strings.TrimSpace(*input.InstitutionID)
		if trimmed == "" {
			return nil, errors.New("institution id cannot be empty")
		}
		input.InstitutionID = &trimmed
	}
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, errors.New("campus name cannot be empty")
		}
		input.Name = &trimmed
	}
	if input.Slug != nil {
		normalized := campusSlugify(*input.Slug)
		if normalized == "" {
			return nil, errors.New("campus slug cannot be empty")
		}
		input.Slug = &normalized
	}
	if input.Status != nil {
		normalizedStatus := strings.TrimSpace(*input.Status)
		if !isCampusStatusValid(normalizedStatus) {
			return nil, fmt.Errorf("invalid campus status %q", normalizedStatus)
		}
		input.Status = &normalizedStatus
	}

	return s.repo.Update(ctx, id, input)
}

// Delete removes a campus record.
func (s *CampusService) Delete(ctx context.Context, id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, errors.New("campus id is required")
	}
	return s.repo.Delete(ctx, id)
}

func campusSlugify(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	trimmed = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(trimmed, "-")
	trimmed = strings.Trim(trimmed, "-")
	return trimmed
}

func isCampusStatusValid(status string) bool {
	switch strings.ToLower(status) {
	case "active", "inactive", "suspended":
		return true
	default:
		return false
	}
}
