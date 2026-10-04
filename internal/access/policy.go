package access

import "corporate-workspace/internal/domain"

func CanManageDirectory(role domain.Role) bool {
	return role == domain.RoleSystemAdmin ||
		role == domain.RoleCompanyAdmin ||
		role == domain.RoleHR
}

func CanSeeArchived(role domain.Role) bool {
	return role == domain.RoleSystemAdmin ||
		role == domain.RoleCompanyAdmin ||
		role == domain.RoleHR
}

func CanEditOwnContacts(actorEmployeeID, targetEmployeeID int) bool {
	return actorEmployeeID == targetEmployeeID
}
