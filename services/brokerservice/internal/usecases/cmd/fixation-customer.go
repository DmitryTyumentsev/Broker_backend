package cmd

type FixedBy string
type CustomerID string
type FixFor string
type BrokerID string

type FixationCustomerRequest struct {
	BrokerID   BrokerID
	CustomerID CustomerID //по какому правилу надо создавать отдельно тип для поля? ок ли везде пихать? где есть смысл их делать и а где нет?
	FixFor     FixFor
	FixedBy    FixedBy
}

type FixationCustomerResponse struct {
	FixationID string
	Status     string
	FixedAt    string
	ExpiresAt  string
}
