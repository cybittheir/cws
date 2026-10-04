package repository

import (
	"context"
	"corporate-workspace/internal/domain"
)

type Store interface {
	Authenticate(context.Context, string, string) (domain.User, error)
	UserByID(context.Context, int) (domain.User, error)
	SetActiveCompany(context.Context, int, int) error
	UpdateSessionCompany(context.Context, string, int) error
	Directory(context.Context, int, string) ([]domain.Employee, []domain.Department, int, int, error)
	Employee(context.Context, int, int) (domain.Employee, error)
	UpdateOwnContacts(context.Context, int, int, ContactUpdate) (domain.Employee, error)
	CreateSession(context.Context, int, int, string) error
	Session(context.Context, string) (domain.User, error)
	DeleteSession(context.Context, string) error
	Close() error
}
type ContactUpdate struct {
	MobilePhone, PersonalEmail, TelegramURL, MattermostURL string
	ShowPersonalEmail                                      bool
	Version                                                int
}
