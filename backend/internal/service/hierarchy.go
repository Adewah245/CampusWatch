package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"CampusWatch/backend/internal/model"
)

type ClusterRepository interface {
	Create(context.Context, model.Cluster) (*model.Cluster, error)
	List(context.Context) ([]model.Cluster, error)
	GetByID(context.Context, string) (*model.Cluster, error)
	Update(context.Context, string, model.UpdateClusterRequest) (*model.Cluster, error)
	Delete(context.Context, string) (bool, error)
}

type ClusterService struct{ repo ClusterRepository }

func NewClusterService(repo ClusterRepository) *ClusterService { return &ClusterService{repo: repo} }

func (s *ClusterService) Create(ctx context.Context, campusID, name, slug, status string) (*model.Cluster, error) {
	campusID = strings.TrimSpace(campusID)
	if campusID == "" {
		return nil, errors.New("campus id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("cluster name is required")
	}
	slug = hierarchySlugify(slug)
	if slug == "" {
		slug = hierarchySlugify(name)
	}
	if slug == "" {
		return nil, errors.New("cluster slug is required")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "active"
	}
	if !hierarchyStatusValid(status) {
		return nil, fmt.Errorf("invalid cluster status %q", status)
	}
	return s.repo.Create(ctx, model.Cluster{CampusID: campusID, Name: name, Slug: slug, Status: status})
}

func (s *ClusterService) List(ctx context.Context) ([]model.Cluster, error) { return s.repo.List(ctx) }

func (s *ClusterService) GetByID(ctx context.Context, id string) (*model.Cluster, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("cluster id is required")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *ClusterService) Update(ctx context.Context, id string, input model.UpdateClusterRequest) (*model.Cluster, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("cluster id is required")
	}
	if input.CampusID != nil {
		value := strings.TrimSpace(*input.CampusID)
		if value == "" {
			return nil, errors.New("campus id cannot be empty")
		}
		input.CampusID = &value
	}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if value == "" {
			return nil, errors.New("cluster name cannot be empty")
		}
		input.Name = &value
	}
	if input.Slug != nil {
		value := hierarchySlugify(*input.Slug)
		if value == "" {
			return nil, errors.New("cluster slug cannot be empty")
		}
		input.Slug = &value
	}
	if input.Status != nil {
		value := strings.TrimSpace(*input.Status)
		if !hierarchyStatusValid(value) {
			return nil, fmt.Errorf("invalid cluster status %q", value)
		}
		input.Status = &value
	}
	return s.repo.Update(ctx, id, input)
}

func (s *ClusterService) Delete(ctx context.Context, id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, errors.New("cluster id is required")
	}
	return s.repo.Delete(ctx, id)
}

type LocationRepository interface {
	Create(context.Context, model.Location) (*model.Location, error)
	List(context.Context) ([]model.Location, error)
	GetByID(context.Context, string) (*model.Location, error)
	Update(context.Context, string, model.UpdateLocationRequest) (*model.Location, error)
	Delete(context.Context, string) (bool, error)
}

type LocationService struct{ repo LocationRepository }

func NewLocationService(repo LocationRepository) *LocationService {
	return &LocationService{repo: repo}
}

func (s *LocationService) Create(ctx context.Context, clusterID, name, slug, locationType, status string) (*model.Location, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, errors.New("cluster id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("location name is required")
	}
	slug = hierarchySlugify(slug)
	if slug == "" {
		slug = hierarchySlugify(name)
	}
	if slug == "" {
		return nil, errors.New("location slug is required")
	}
	locationType = strings.TrimSpace(locationType)
	if locationType == "" {
		locationType = "table"
	}
	if !locationTypeValid(locationType) {
		return nil, fmt.Errorf("invalid location type %q", locationType)
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "active"
	}
	if !hierarchyStatusValid(status) {
		return nil, fmt.Errorf("invalid location status %q", status)
	}
	return s.repo.Create(ctx, model.Location{ClusterID: clusterID, Name: name, Slug: slug, LocationType: locationType, Status: status})
}

func (s *LocationService) List(ctx context.Context) ([]model.Location, error) {
	return s.repo.List(ctx)
}

func (s *LocationService) GetByID(ctx context.Context, id string) (*model.Location, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("location id is required")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *LocationService) Update(ctx context.Context, id string, input model.UpdateLocationRequest) (*model.Location, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("location id is required")
	}
	if input.ClusterID != nil {
		value := strings.TrimSpace(*input.ClusterID)
		if value == "" {
			return nil, errors.New("cluster id cannot be empty")
		}
		input.ClusterID = &value
	}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if value == "" {
			return nil, errors.New("location name cannot be empty")
		}
		input.Name = &value
	}
	if input.Slug != nil {
		value := hierarchySlugify(*input.Slug)
		if value == "" {
			return nil, errors.New("location slug cannot be empty")
		}
		input.Slug = &value
	}
	if input.LocationType != nil {
		value := strings.TrimSpace(*input.LocationType)
		if !locationTypeValid(value) {
			return nil, fmt.Errorf("invalid location type %q", value)
		}
		input.LocationType = &value
	}
	if input.Status != nil {
		value := strings.TrimSpace(*input.Status)
		if !hierarchyStatusValid(value) {
			return nil, fmt.Errorf("invalid location status %q", value)
		}
		input.Status = &value
	}
	return s.repo.Update(ctx, id, input)
}

func (s *LocationService) Delete(ctx context.Context, id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, errors.New("location id is required")
	}
	return s.repo.Delete(ctx, id)
}

func hierarchySlugify(value string) string {
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(strings.ToLower(strings.TrimSpace(value)), "-")
	return strings.Trim(value, "-")
}

func hierarchyStatusValid(value string) bool {
	switch strings.ToLower(value) {
	case "active", "inactive", "suspended":
		return true
	default:
		return false
	}
}

func locationTypeValid(value string) bool {
	switch strings.ToLower(value) {
	case "table", "room", "lab", "office":
		return true
	default:
		return false
	}
}
