package roles

const (
	SuperAdmin       = "superadmin"
	DeveloperAdmin   = "developer_admin"
	AccountManager   = "account_manager"
	SalesManager     = "sales_manager"
	AgencyOwner      = "agency_owner"
	BrokerTeamLead   = "broker_team_lead"
	BrokerTeamMember = "broker_team_member"
)

var All = []string{
	SuperAdmin,
	DeveloperAdmin,
	AccountManager,
	SalesManager,
	AgencyOwner,
	BrokerTeamLead,
	BrokerTeamMember,
}

func Known(role string) bool {
	for _, known := range All {
		if role == known {
			return true
		}
	}

	return false
}
