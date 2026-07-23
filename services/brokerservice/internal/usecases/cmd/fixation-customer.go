package cmd

type FixationCustomerRequest struct {
	BrokerID   string
	CustomerID string
	FixFor     string
	FixedBy    string
}
