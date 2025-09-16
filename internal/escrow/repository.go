package escrow

import "context"

type SingleRepository interface {
	CreateOrUpdate(ctx context.Context, e SingleReleaseJSON) error
	Get(ctx context.Context, contractID string) (map[string]any, error)
	Delete(ctx context.Context, contractID string) error
}

type MultiRepository interface {
	CreateOrUpdate(ctx context.Context, e MultiReleaseJSON) error
	Get(ctx context.Context, contractID string) (map[string]any, error)
	Delete(ctx context.Context, contractID string) error
}
