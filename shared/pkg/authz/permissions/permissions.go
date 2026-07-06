package permissions

const (
	APIProtectedAccess   = "api.protected.access"
	DeveloperAdminAccess = "developer.admin.access"

	CustomerFixationCreate = "customer_fixation.create"
	CustomerFixationRead   = "customer_fixation.read"
	CustomerFixationManage = "customer_fixation.manage"
)

var All = []string{
	APIProtectedAccess,
	DeveloperAdminAccess,
	CustomerFixationCreate,
	CustomerFixationRead,
	CustomerFixationManage,
}

func Known(permission string) bool {
	for _, known := range All {
		if permission == known {
			return true
		}
	}

	return false
}
