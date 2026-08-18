package entity

type UserRole string

const (
	RoleSuperadmin       UserRole = "superadmin"
	RoleDeveloperAdmin   UserRole = "developer_admin"
	RoleAccountManager   UserRole = "account_manager"
	RoleSalesManager     UserRole = "sales_manager"
	RoleAgencyOwner      UserRole = "agency_owner"
	RoleBrokerTeamLead   UserRole = "broker_team_lead"
	RoleBrokerTeamMember UserRole = "broker_team_member"
)
