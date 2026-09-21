package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"CampusWatch/backend/internal/model"
)

// InstitutionRepository defines the persistence contract for institutions.
type InstitutionRepository interface {
	Create(ctx context.Context, institution model.Institution) (*model.Institution, error)
	List(ctx context.Context) ([]model.Institution, error)
	GetByID(ctx context.Context, id string) (*model.Institution, error)
	FindBySlug(ctx context.Context, slug string) (*model.Institution, error)
	Update(ctx context.Context, id string, input model.UpdateInstitutionRequest) (*model.Institution, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// InstitutionService provides the business logic for institution operations.
type InstitutionService struct {
	repo InstitutionRepository
}

// NewInstitutionService creates an institution service with a repository implementation.
func NewInstitutionService(repo InstitutionRepository) *InstitutionService {
	return &InstitutionService{repo: repo}
}

// Create creates a new institution after validating key fields.
func (s *InstitutionService) Create(ctx context.Context, name, slug, status string) (*model.Institution, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, errors.New("institution name is required")
	}

	normalizedSlug := slugify(slug)
	if normalizedSlug == "" {
		normalizedSlug = slugify(trimmedName)
	}
	if normalizedSlug == "" {
		return nil, errors.New("institution slug is required")
	}

	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus == "" {
		normalizedStatus = "active"
	}
	if !isValidInstitutionStatus(normalizedStatus) {
		return nil, fmt.Errorf("invalid institution status %q", normalizedStatus)
	}

	institution := model.Institution{
		Name:   trimmedName,
		Slug:   normalizedSlug,
		Status: normalizedStatus,
	}

	return s.repo.Create(ctx, institution)
}

// List fetches all institution records.
func (s *InstitutionService) List(ctx context.Context) ([]model.Institution, error) {
	return s.repo.List(ctx)
}

// GetByID loads one institution by ID.
func (s *InstitutionService) GetByID(ctx context.Context, id string) (*model.Institution, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("institution id is required")
	}
	return s.repo.GetByID(ctx, id)
}

// GetBySlug loads one institution by slug, or by name when no slug is given.
//
// The input is slugified with the same rule Create uses, so a caller can pass
// either "Adewah University" or "adewah-university" and reach the same row.
// Returns (nil, nil) when nothing matches, leaving "create it" to the caller.
func (s *InstitutionService) GetBySlug(ctx context.Context, slugOrName string) (*model.Institution, error) {
	normalized := slugify(slugOrName)
	if normalized == "" {
		return nil, errors.New("institution slug is required")
	}
	return s.repo.FindBySlug(ctx, normalized)
}

// Update applies partial field updates to an institution.
func (s *InstitutionService) Update(ctx context.Context, id string, input model.UpdateInstitutionRequest) (*model.Institution, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("institution id is required")
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, errors.New("institution name cannot be empty")
		}
		input.Name = &trimmed
	}
	if input.Slug != nil {
		normalized := slugify(*input.Slug)
		if normalized == "" {
			return nil, errors.New("institution slug cannot be empty")
		}
		input.Slug = &normalized
	}
	if input.Status != nil {
		normalizedStatus := strings.TrimSpace(*input.Status)
		if !isValidInstitutionStatus(normalizedStatus) {
			return nil, fmt.Errorf("invalid institution status %q", normalizedStatus)
		}
		input.Status = &normalizedStatus
	}

	return s.repo.Update(ctx, id, input)
}

// Delete removes an institution record.
func (s *InstitutionService) Delete(ctx context.Context, id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, errors.New("institution id is required")
	}
	return s.repo.Delete(ctx, id)
}

func slugify(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	trimmed = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(trimmed, "-")
	trimmed = strings.Trim(trimmed, "-")
	return trimmed
}

func isValidInstitutionStatus(status string) bool {
	switch strings.ToLower(status) {
	case "active", "inactive", "suspended":
		return true
	default:
		return false
	}
}
