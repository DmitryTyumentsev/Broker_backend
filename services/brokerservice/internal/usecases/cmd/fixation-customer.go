package cmd

type BrokerID string
type CustomerID string
type ManagerID string
type ConnectCustomerRequest struct {
	CustomerID CustomerID
	BrokerID   BrokerID
	ManagerID  ManagerID
}

type FullNameManager struct {
	ManagerLastName   string
	ManagerFirstName  string
	ManagerMiddleName string
}
