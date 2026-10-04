package memory

import (
	"context"
	"corporate-workspace/internal/domain/directory"
	"corporate-workspace/internal/domain/identity"
	"corporate-workspace/internal/repository"
	"errors"
	"sort"
	"strings"
	"time"
)

type Store struct {
	users       map[int]identity.User
	employees   map[int]directory.Employee
	departments map[int]directory.Department
	audits      []repository.AuditEntry
}

func New() *Store {
	s := &Store{users: map[int]identity.User{}, employees: map[int]directory.Employee{}, departments: map[int]directory.Department{}}
	s.seed()
	return s
}
func (s *Store) seed() {
	s.departments[1] = directory.Department{ID: 1, CompanyID: 1, Name: "ИТ-служба", Path: "Дирекция → ИТ-служба", SortOrder: 1}
	s.departments[2] = directory.Department{ID: 2, CompanyID: 1, Name: "Отдел продаж", Path: "Коммерческий блок → Отдел продаж", SortOrder: 2}
	s.departments[3] = directory.Department{ID: 3, CompanyID: 2, Name: "Разработка", Path: "Технологии → Разработка", SortOrder: 1}
	s.users[1] = identity.User{ID: 1, Email: "admin@alpha.example", PasswordHash: "demo", PersonName: "Анна Петрова", ActiveCompanyID: 1, Memberships: []identity.Membership{{1, "ООО «Альфа»", "company_admin", true}, {2, "ООО «Бета»", "department_manager", true}}, Theme: "mist"}
	s.users[2] = identity.User{ID: 2, Email: "ivan@alpha.example", PasswordHash: "demo", PersonName: "Иван Петров", ActiveCompanyID: 1, Memberships: []identity.Membership{{1, "ООО «Альфа»", "employee", true}}, Theme: "sage"}
	hired := time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC)
	s.employees[1] = directory.Employee{ID: 1, CompanyID: 1, FirstName: "Анна", LastName: "Петрова", Position: "HR-менеджер", City: "Владивосток", InternalPhone: "101", MobilePhone: "+7 900 000-00-01", WorkEmail: "admin@alpha.example", TelegramURL: "https://t.me/anna_demo", DepartmentIDs: []int{1}, PrimaryDepartmentID: 1, Role: directory.RoleCompanyAdmin, Status: directory.StatusActive, DateHired: hired, Version: 1}
	s.employees[2] = directory.Employee{ID: 2, CompanyID: 1, FirstName: "Иван", LastName: "Петров", MiddleName: "Алексеевич", Position: "Senior Go-разработчик", City: "Владивосток", InternalPhone: "205", MobilePhone: "+7 900 000-00-02", WorkEmail: "ivan@alpha.example", TelegramURL: "https://t.me/ivan_demo", MattermostURL: "https://chat.example/direct/ivan", DepartmentIDs: []int{1}, PrimaryDepartmentID: 1, Role: directory.RoleEmployee, Status: directory.StatusActive, DateHired: hired, Version: 1}
	s.employees[3] = directory.Employee{ID: 3, CompanyID: 1, FirstName: "Мария", LastName: "Соколова", Position: "Аналитик", City: "Владивосток", InternalPhone: "206", MobilePhone: "+7 900 000-00-03", WorkEmail: "m.sokolova@alpha.example", DepartmentIDs: []int{1, 2}, PrimaryDepartmentID: 1, Role: directory.RoleEmployee, Status: directory.StatusActive, DateHired: hired, Version: 1}
	s.employees[4] = directory.Employee{ID: 4, CompanyID: 1, FirstName: "Ольга", LastName: "Волкова", Position: "Руководитель отдела продаж", City: "Владивосток", InternalPhone: "311", MobilePhone: "+7 900 000-00-04", WorkEmail: "o.volkova@alpha.example", MattermostURL: "https://chat.example/direct/olga", DepartmentIDs: []int{2}, PrimaryDepartmentID: 2, Role: directory.RoleDepartmentManager, Status: directory.StatusActive, DateHired: hired, Version: 1}
	s.employees[5] = directory.Employee{ID: 5, CompanyID: 1, FirstName: "Сергей", LastName: "Ким", Position: "Менеджер по продажам", City: "Владивосток", InternalPhone: "312", MobilePhone: "+7 900 000-00-05", WorkEmail: "s.kim@alpha.example", DepartmentIDs: []int{2}, PrimaryDepartmentID: 2, Role: directory.RoleEmployee, Status: directory.StatusActive, DateHired: hired, Version: 1}
}
func (s *Store) Authenticate(_ context.Context, email, password string) (identity.User, error) {
	for _, u := range s.users {
		if strings.EqualFold(u.Email, email) && password == "demo" {
			return u, nil
		}
	}
	return identity.User{}, errors.New("invalid credentials")
}
func (s *Store) UserByID(_ context.Context, id int) (identity.User, error) {
	u, ok := s.users[id]
	if !ok {
		return identity.User{}, errors.New("user not found")
	}
	return u, nil
}
func (s *Store) CompaniesForUser(ctx context.Context, id int) ([]identity.Membership, error) {
	u, e := s.UserByID(ctx, id)
	return u.Memberships, e
}
func (s *Store) Directory(_ context.Context, companyID int, f repository.DirectoryFilter) (repository.DirectoryResult, error) {
	var rows []directory.Employee
	cities := map[string]bool{}
	for _, e := range s.employees {
		if e.CompanyID != companyID {
			continue
		}
		if e.Status != directory.StatusActive && !f.IncludeArchived {
			continue
		}
		q := strings.ToLower(f.Query)
		if q != "" && !strings.Contains(strings.ToLower(e.FullName()+" "+e.Position), q) {
			continue
		}
		if f.DepartmentID > 0 && !contains(e.DepartmentIDs, f.DepartmentID) {
			continue
		}
		rows = append(rows, e)
		if e.Status == directory.StatusActive && e.City != "" {
			cities[e.City] = true
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].LastName < rows[j].LastName })
	var deps []directory.Department
	for _, d := range s.departments {
		if d.CompanyID == companyID {
			deps = append(deps, d)
		}
	}
	sort.Slice(deps, func(i, j int) bool { return deps[i].SortOrder < deps[j].SortOrder })
	return repository.DirectoryResult{Employees: rows, Departments: deps, Total: len(rows), DistinctCities: len(cities)}, nil
}
func (s *Store) EmployeeByID(_ context.Context, companyID, id int) (directory.Employee, error) {
	e, ok := s.employees[id]
	if !ok || e.CompanyID != companyID {
		return directory.Employee{}, errors.New("employee not found")
	}
	return e, nil
}
func (s *Store) UpdateOwnContacts(_ context.Context, userID, employeeID int, u repository.ContactUpdate) (directory.Employee, error) {
	e, ok := s.employees[employeeID]
	if !ok || e.ID != userID {
		return directory.Employee{}, errors.New("forbidden")
	}
	if e.Version != u.Version {
		return directory.Employee{}, errors.New("version conflict")
	}
	e.MobilePhone = u.MobilePhone
	e.PersonalEmail = u.PersonalEmail
	e.ShowPersonalEmail = u.ShowPersonalEmail
	e.TelegramURL = u.TelegramURL
	e.MattermostURL = u.MattermostURL
	e.Version++
	s.employees[employeeID] = e
	return e, nil
}
func (s *Store) Audit(_ context.Context, e repository.AuditEntry) error {
	s.audits = append(s.audits, e)
	return nil
}
func contains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
