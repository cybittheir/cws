package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"corporate-workspace/internal/domain"
	"corporate-workspace/internal/repository"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if _, err := db.Exec("PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.seed(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	migration, err := migrationFS.ReadFile("migrations/000001_init.sql")
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, string(migration))
	return err
}

func (s *Store) seed(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
INSERT INTO companies (id, name, logo_text, time_zone, default_theme) VALUES
(1, 'ООО «Альфа»', 'A', 'Asia/Vladivostok', 'mist'),
(2, 'ООО «Бета»', 'B', 'Asia/Vladivostok', 'sage');

INSERT INTO departments (id, company_id, path, sort_order) VALUES
(1, 1, 'Дирекция → ИТ-служба', 1),
(2, 1, 'Коммерческий блок → Отдел продаж', 2),
(3, 2, 'Технологии → Разработка', 1);
`)
	if err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("demo"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO users (id, email, password_hash, full_name, theme) VALUES
(1, ?, ?, ?, ?),
(2, ?, ?, ?, ?);
`,
		"admin@alpha.example", string(passwordHash), "Анна Петрова", "mist",
		"ivan@alpha.example", string(passwordHash), "Иван Петров", "sage",
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO company_memberships (user_id, company_id, role, status) VALUES
(1, 1, 'company_admin', 'active'),
(1, 2, 'department_manager', 'active'),
(2, 1, 'employee', 'active');

INSERT INTO employees (
    id, user_id, company_id, first_name, last_name, middle_name,
    position, city, internal_phone, mobile_phone, work_email,
    primary_department_id, role, status, date_hired, version
) VALUES
(1, 1, 1, 'Анна', 'Петрова', '', 'HR-менеджер', 'Владивосток', '101', '+7 900 000-00-01', 'admin@alpha.example', 1, 'company_admin', 'active', '2021-03-01', 1),
(2, 2, 1, 'Иван', 'Петров', 'Алексеевич', 'Senior Go-разработчик', 'Владивосток', '205', '+7 900 000-00-02', 'ivan@alpha.example', 1, 'employee', 'active', '2021-03-01', 1),
(3, NULL, 1, 'Мария', 'Соколова', '', 'Аналитик', 'Владивосток', '206', '+7 900 000-00-03', 'm.sokolova@alpha.example', 1, 'employee', 'active', '2021-03-01', 1),
(4, NULL, 1, 'Ольга', 'Волкова', '', 'Руководитель отдела продаж', 'Владивосток', '311', '+7 900 000-00-04', 'o.volkova@alpha.example', 2, 'department_manager', 'active', '2021-03-01', 1),
(5, NULL, 1, 'Сергей', 'Ким', '', 'Менеджер по продажам', 'Владивосток', '312', '+7 900 000-00-05', 's.kim@alpha.example', 2, 'employee', 'active', '2021-03-01', 1);

INSERT INTO employee_departments (employee_id, department_id, is_primary) VALUES
(1, 1, 1),
(2, 1, 1),
(3, 1, 1),
(3, 2, 0),
(4, 2, 1),
(5, 2, 1);
`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
	var userID int
	var passwordHash string

	err := s.db.QueryRowContext(ctx, `
SELECT id, password_hash
FROM users
WHERE lower(email) = lower(?) AND status = 'active'
`, email).Scan(&userID, &passwordHash)
	if err != nil {
		return domain.User{}, errors.New("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		return domain.User{}, errors.New("invalid credentials")
	}

	return s.user(ctx, userID)
}

func (s *Store) UserByID(ctx context.Context, userID int) (domain.User, error) {
	return s.user(ctx, userID)
}

func (s *Store) user(ctx context.Context, userID int) (domain.User, error) {
	var user domain.User

	err := s.db.QueryRowContext(ctx, `
SELECT id, email, full_name, theme
FROM users
WHERE id = ? AND status = 'active'
`, userID).Scan(&user.ID, &user.Email, &user.Name, &user.Theme)
	if err != nil {
		return domain.User{}, err
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT m.company_id, c.name, m.role, m.status
FROM company_memberships AS m
JOIN companies AS c ON c.id = m.company_id
WHERE m.user_id = ?
ORDER BY m.company_id
`, userID)
	if err != nil {
		return domain.User{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var membership domain.Membership
		var status string
		if err := rows.Scan(
			&membership.CompanyID,
			&membership.CompanyName,
			&membership.Role,
			&status,
		); err != nil {
			return domain.User{}, err
		}
		membership.Active = status == "active"
		user.Memberships = append(user.Memberships, membership)
		if user.ActiveCompanyID == 0 && membership.Active {
			user.ActiveCompanyID = membership.CompanyID
		}
	}
	if err := rows.Err(); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *Store) SetActiveCompany(ctx context.Context, userID, companyID int) error {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT count(*)
FROM company_memberships
WHERE user_id = ? AND company_id = ? AND status = 'active'
`, userID, companyID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("company access denied")
	}
	return nil
}

func (s *Store) UpdateSessionCompany(ctx context.Context, token string, companyID int) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE sessions
SET active_company_id = ?
WHERE token = ? AND expires_at > CURRENT_TIMESTAMP
`, companyID, token)
	return err
}

func (s *Store) Directory(ctx context.Context, companyID int, query string) ([]domain.Employee, []domain.Department, int, int, error) {
	like := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"

	rows, err := s.db.QueryContext(ctx, `
SELECT
    id, company_id, first_name, last_name, middle_name, position, city,
    internal_phone, mobile_phone, work_email,
    coalesce(personal_email, ''), show_personal_email,
    coalesce(telegram_url, ''), coalesce(mattermost_url, ''),
    primary_department_id, role, version, date_hired
FROM employees
WHERE company_id = ?
  AND status = 'active'
  AND lower(last_name || ' ' || first_name || ' ' || middle_name || ' ' || position) LIKE ?
ORDER BY last_name, first_name
`, companyID, like)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	defer rows.Close()

	var employees []domain.Employee
	for rows.Next() {
		var employee domain.Employee
		var dateHired string
		if err := rows.Scan(
			&employee.ID,
			&employee.CompanyID,
			&employee.FirstName,
			&employee.LastName,
			&employee.MiddleName,
			&employee.Position,
			&employee.City,
			&employee.InternalPhone,
			&employee.MobilePhone,
			&employee.WorkEmail,
			&employee.PersonalEmail,
			&employee.ShowPersonalEmail,
			&employee.TelegramURL,
			&employee.MattermostURL,
			&employee.PrimaryDepartmentID,
			&employee.Role,
			&employee.Version,
			&dateHired,
		); err != nil {
			return nil, nil, 0, 0, err
		}
		employee.DateHired, _ = time.Parse("2006-01-02", dateHired)
		employees = append(employees, employee)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, 0, 0, err
	}

	departmentRows, err := s.db.QueryContext(ctx, `
SELECT id, company_id, path, sort_order
FROM departments
WHERE company_id = ?
ORDER BY sort_order
`, companyID)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	defer departmentRows.Close()

	var departments []domain.Department
	for departmentRows.Next() {
		var department domain.Department
		if err := departmentRows.Scan(
			&department.ID,
			&department.CompanyID,
			&department.Path,
			&department.SortOrder,
		); err != nil {
			return nil, nil, 0, 0, err
		}
		departments = append(departments, department)
	}
	if err := departmentRows.Err(); err != nil {
		return nil, nil, 0, 0, err
	}

	var distinctCities int
	var total int
	err = s.db.QueryRowContext(ctx, `
SELECT count(DISTINCT city), count(*)
FROM employees
WHERE company_id = ? AND status = 'active'
`, companyID).Scan(&distinctCities, &total)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	return employees, departments, total, distinctCities, nil
}

func (s *Store) Employee(ctx context.Context, companyID, employeeID int) (domain.Employee, error) {
	employees, _, _, _, err := s.Directory(ctx, companyID, "")
	if err != nil {
		return domain.Employee{}, err
	}
	for _, employee := range employees {
		if employee.ID == employeeID {
			return employee, nil
		}
	}
	return domain.Employee{}, sql.ErrNoRows
}

func (s *Store) UpdateOwnContacts(
	ctx context.Context,
	userID, employeeID int,
	update repository.ContactUpdate,
) (domain.Employee, error) {
	result, err := s.db.ExecContext(ctx, `
UPDATE employees
SET
    mobile_phone = ?,
    personal_email = ?,
    show_personal_email = ?,
    telegram_url = ?,
    mattermost_url = ?,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ? AND version = ?
`,
		update.MobilePhone,
		update.PersonalEmail,
		update.ShowPersonalEmail,
		update.TelegramURL,
		update.MattermostURL,
		employeeID,
		userID,
		update.Version,
	)
	if err != nil {
		return domain.Employee{}, err
	}

	changed, err := result.RowsAffected()
	if err != nil {
		return domain.Employee{}, err
	}
	if changed == 0 {
		return domain.Employee{}, errors.New("version conflict or forbidden")
	}

	var companyID int
	if err := s.db.QueryRowContext(ctx, "SELECT company_id FROM employees WHERE id = ?", employeeID).Scan(&companyID); err != nil {
		return domain.Employee{}, err
	}

	_, err = s.db.ExecContext(ctx, `
INSERT INTO audit_logs (actor_user_id, company_id, action, entity_type, entity_id, details)
VALUES (?, ?, ?, ?, ?, ?)
`, userID, companyID, "contacts.updated", "employee", employeeID, "self-service update")
	if err != nil {
		return domain.Employee{}, err
	}

	return s.Employee(ctx, companyID, employeeID)
}

func (s *Store) CreateSession(ctx context.Context, userID, companyID int, token string) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO sessions (token, user_id, active_company_id, expires_at)
VALUES (?, ?, ?, ?)
`, token, userID, companyID, time.Now().UTC().Add(8*time.Hour))
	return err
}

func (s *Store) Session(ctx context.Context, token string) (domain.User, error) {
	var userID int
	var activeCompanyID int

	err := s.db.QueryRowContext(ctx, `
SELECT user_id, active_company_id
FROM sessions
WHERE token = ? AND expires_at > CURRENT_TIMESTAMP
`, token).Scan(&userID, &activeCompanyID)
	if err != nil {
		return domain.User{}, err
	}

	user, err := s.user(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	user.ActiveCompanyID = activeCompanyID
	return user, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE token = ?", token)
	return err
}

func NewToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
