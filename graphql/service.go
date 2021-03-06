package graphql

import (
	"github.com/samsarahq/thunder/graphql/schemabuilder"

	"github.com/Depado/thundermonit/domain"
)

const (
	serviceName = "service"
	serviceDesc = "A type that describes a single service"
)

// Simple query with a single int64 ID inside
type idQuery struct {
	ID int64 `graphql:"id"`
}

// RegisterService registers the Service struct and methods to the schema
func (r RequestHandler) RegisterService(s *schemabuilder.Schema) {
	// Object registration
	obj := s.Object(serviceName, domain.Service{})
	obj.Description = serviceDesc
	obj.FieldFunc("repo", func(s *domain.Service) (*domain.Repo, error) {
		return r.Interactor.GetRepo(s)
	})

	// Queries registration
	q := s.Query()
	q.FieldFunc("service", func(args idQuery) (*domain.Service, error) {
		return r.Interactor.GetService(args.ID)
	})
	q.FieldFunc("services", func() ([]*domain.Service, error) {
		return r.Interactor.GetAllServices()
	})
}
