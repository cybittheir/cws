package service

import (
	"context"
	"corporate-workspace/internal/domain"
	"corporate-workspace/internal/repository"
)

type DirectoryService struct{ store repository.Store }

func NewDirectory(s repository.Store) *DirectoryService { return &DirectoryService{s} }
func (s *DirectoryService) List(c context.Context, company int, q string) ([]domain.Employee, []domain.Department, int, int, error) {
	return s.store.Directory(c, company, q)
}
func (s *DirectoryService) Employee(c context.Context, company, id int) (domain.Employee, error) {
	return s.store.Employee(c, company, id)
}
func (s *DirectoryService) Update(c context.Context, user, employee int, u repository.ContactUpdate) (domain.Employee, error) {
	return s.store.UpdateOwnContacts(c, user, employee, u)
}
