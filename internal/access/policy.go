package access

import "corporate-workspace/internal/domain/directory"

func CanManageDirectory(role directory.Role) bool {
	return role == directory.RoleSystemAdmin || role == directory.RoleCompanyAdmin || role == directory.RoleHR
}
func CanSeeArchived(role directory.Role) bool {
	return role == directory.RoleSystemAdmin || role == directory.RoleCompanyAdmin || role == directory.RoleHR
}
func CanEditOwnContacts(actorID, targetID int) bool { return actorID == targetID }
