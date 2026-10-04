package service

import (
	"context"
	"corporate-workspace/internal/domain/directory"
	"corporate-workspace/internal/repository"
)

type DirectoryService struct{ store repository.Store }

func NewDirectory(store repository.Store) *DirectoryService { return &DirectoryService{store} }
func (s *DirectoryService) List(ctx context.Context, companyID int, f repository.DirectoryFilter) (repository.DirectoryResult, error) {
	if f.PerPage <= 0 {
		f.PerPage = 50
	}
	return s.store.Directory(ctx, companyID, f)
}
func (s *DirectoryService) Employee(ctx context.Context, companyID, id int) (directory.Employee, error) {
	return s.store.EmployeeByID(ctx, companyID, id)
}
func (s *DirectoryService) UpdateOwnContacts(ctx context.Context, userID, employeeID int, u repository.ContactUpdate) (directory.Employee, error) {
	e, err := s.store.UpdateOwnContacts(ctx, userID, employeeID, u)
	if err == nil {
		_ = s.store.Audit(ctx, repository.AuditEntry{ActorUserID: userID, CompanyID: e.CompanyID, Action: "contacts.updated", EntityType: "employee", EntityID: employeeID})
	}
	return e, err
}
