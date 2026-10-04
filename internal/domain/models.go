package domain

import "time"

type Role string

const (
	RoleSystemAdmin       Role = "system_admin"
	RoleCompanyAdmin      Role = "company_admin"
	RoleHR                Role = "hr"
	RoleDepartmentManager Role = "department_manager"
	RoleEmployee          Role = "employee"
)

type User struct {
	ID              int
	Email           string
	Name            string
	ActiveCompanyID int
	Theme           string
	Memberships     []Membership
}
type Membership struct {
	CompanyID   int
	CompanyName string
	Role        Role
	Active      bool
}
type Employee struct {
	ID                                                                                                                                int
	CompanyID                                                                                                                         int
	FirstName, LastName, MiddleName, Position, City, InternalPhone, MobilePhone, WorkEmail, PersonalEmail, TelegramURL, MattermostURL string
	ShowPersonalEmail                                                                                                                 bool
	PrimaryDepartmentID                                                                                                               int
	Role                                                                                                                              Role
	Version                                                                                                                           int
	DateHired                                                                                                                         time.Time
}

func (e Employee) FullName() string {
	if e.MiddleName == "" {
		return e.LastName + " " + e.FirstName
	}
	return e.LastName + " " + e.FirstName + " " + e.MiddleName
}

type Department struct {
	ID        int
	CompanyID int
	Path      string
	SortOrder int
}
