package permissions

const (
	CustomerFixationCreate = "customer_fixation.create"
	CustomerFixationRead   = "customer_fixation.read"
	CustomerFixationCancel = "customer_fixation.cancel"

	BrokerTeamManage = "broker_team.manage"

	ProjectRead   = "project.read"
	ProjectManage = "project.manage"

	LotRead = "lot.read"

	DealRead   = "deal.read"
	DealManage = "deal.manage"
)

var All = []string{
	CustomerFixationCreate,
	CustomerFixationRead,
	CustomerFixationCancel,
	BrokerTeamManage,
	ProjectRead,
	ProjectManage,
	LotRead,
	DealRead,
	DealManage,
}
