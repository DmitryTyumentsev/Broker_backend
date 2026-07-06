package permissions

const (
	CustomerFixationCreate = "customer_fixation.create"
	CustomerFixationRead   = "customer_fixation.read"
	CustomerFixationCancel = "customer_fixation.cancel"
)

var All = []string{
	CustomerFixationCreate,
	CustomerFixationRead,
	CustomerFixationCancel,
}
