package identity

type User struct {
	ID              int
	Email           string
	PasswordHash    string
	PersonName      string
	ActiveCompanyID int
	SystemAdmin     bool
	Memberships     []Membership
	Theme           string
}
type Membership struct {
	CompanyID   int
	CompanyName string
	Role        string
	Active      bool
}
