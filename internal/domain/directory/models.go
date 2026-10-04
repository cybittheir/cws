package directory

import "time"

type Role string

const (
	RoleSystemAdmin       Role = "system_admin"
	RoleCompanyAdmin      Role = "company_admin"
	RoleHR                Role = "hr"
	RoleDepartmentManager Role = "department_manager"
	RoleEmployee          Role = "employee"
)

type MembershipStatus string

const (
	StatusActive   MembershipStatus = "active"
	StatusBlocked  MembershipStatus = "blocked"
	StatusArchived MembershipStatus = "archived"
)

type Company struct {
	ID           int
	Name         string
	LogoText     string
	TimeZone     string
	DefaultTheme string
}
type Department struct {
	ID        int
	CompanyID int
	ParentID  *int
	Name      string
	Path      string
	SortOrder int
}
type Employee struct {
	ID                  int
	CompanyID           int
	FirstName           string
	LastName            string
	MiddleName          string
	Position            string
	City                string
	InternalPhone       string
	MobilePhone         string
	WorkEmail           string
	PersonalEmail       string
	ShowPersonalEmail   bool
	TelegramURL         string
	MattermostURL       string
	DepartmentIDs       []int
	PrimaryDepartmentID int
	Role                Role
	Status              MembershipStatus
	DateHired           time.Time
	LastWorkDay         *time.Time
	DateDismissed       *time.Time
	Version             int
}

func (e Employee) FullName() string {
	if e.MiddleName == "" {
		return e.LastName + " " + e.FirstName
	}
	return e.LastName + " " + e.FirstName + " " + e.MiddleName
}
