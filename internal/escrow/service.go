package escrow

import "context"

type Service struct {
	single SingleRepository
	multi  MultiRepository
}

func NewService(s SingleRepository, m MultiRepository) *Service {
	return &Service{single: s, multi: m}
}

func (s *Service) UpsertSingle(ctx context.Context, in SingleReleaseJSON) error {
	return s.single.CreateOrUpdate(ctx, in)
}
func (s *Service) UpsertMulti(ctx context.Context, in MultiReleaseJSON) error {
	return s.multi.CreateOrUpdate(ctx, in)
}
func (s *Service) Get(ctx context.Context, contractID string) (map[string]any, error) {
	if j, err := s.single.Get(ctx, contractID); err == nil && j != nil {
		return j, nil
	}
	return s.multi.Get(ctx, contractID)
}
func (s *Service) Delete(ctx context.Context, contractID string) error {
	// Intentamos en ambas tablas
	if err := s.single.Delete(ctx, contractID); err == nil {
		return nil
	}
	return s.multi.Delete(ctx, contractID)
}
