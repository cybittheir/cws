package repository

import (
	"context"
	"corporate-workspace/internal/domain/directory"
	"corporate-workspace/internal/domain/identity"
)

type Store interface {
	Authenticate(context.Context, string, string) (identity.User, error)
	UserByID(context.Context, int) (identity.User, error)
	CompaniesForUser(context.Context, int) ([]identity.Membership, error)
	Directory(context.Context, int, DirectoryFilter) (DirectoryResult, error)
	EmployeeByID(context.Context, int, int) (directory.Employee, error)
	UpdateOwnContacts(context.Context, int, int, ContactUpdate) (directory.Employee, error)
	Audit(context.Context, AuditEntry) error
}
type DirectoryFilter struct {
	Query           string
	DepartmentID    int
	Position        string
	IncludeArchived bool
	Page            int
	PerPage         int
}
type DirectoryResult struct {
	Employees      []directory.Employee
	Departments    []directory.Department
	Total          int
	DistinctCities int
}
type ContactUpdate struct {
	MobilePhone       string
	PersonalEmail     string
	ShowPersonalEmail bool
	TelegramURL       string
	MattermostURL     string
	Version           int
}
type AuditEntry struct {
	ActorUserID int
	CompanyID   int
	Action      string
	EntityType  string
	EntityID    int
	Details     string
}
