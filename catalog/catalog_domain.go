package catalog

import (
	"context"

	"github.io/remast/c4stage/shared/paged"
)

type EntityEnvelope struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Kind        string   `json:"kind"`
	Type        string   `json:"type"`
	Lifecycle   string   `json:"lifecycle"`
	Tags        []string `json:"tags"`
}

type System struct {
	EntityEnvelope
	DependsOn []string `json:"dependsOn"`
}

type Container struct {
	EntityEnvelope
	System       string   `json:"system"`
	ConsumesAPIs []string `json:"consumesAPIs"`
	DependsOn    []string `json:"dependsOn"`
}

type API struct {
	EntityEnvelope
	System string
}

type CatalogRepository interface {
	Reset(ctx context.Context) error

	FindSystems(
		ctx context.Context,
		pageParams *paged.PageParams,
	) ([]System, *paged.Page, error)

	CreateAll(
		ctx context.Context,
		entities []any,
	) error
}
