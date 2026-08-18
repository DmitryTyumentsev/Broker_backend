package permissions

const (
	APIProtectedAccess   = "api.protected.access"
	DeveloperAdminAccess = "developer.admin.access"

	FixationNew = "fixation.new"
)

var All = []string{
	APIProtectedAccess,
	DeveloperAdminAccess,
	FixationNew,
}

func Known(permission string) bool {
	for _, known := range All {
		if permission == known {
			return true
		}
	}

	return false
}
