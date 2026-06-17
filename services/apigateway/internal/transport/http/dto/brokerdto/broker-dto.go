package brokerdto

type customerID string
type brokerID string
type managerID string

type ConnectCustomerRequest struct {
	CustomerID customerID `json:"customer_id"`
	BrokerID   brokerID   `json:"broker_id"`
	ManagerID  managerID  `json:"manager_id"`
}

type ConnectCustomerResponse struct {
	ManagerLastName   string `json:"manager_last_name"`
	ManagerFirstName  string `json:"manager_first_name"`
	ManagerMiddleName string `json:"manager_middle_name"`
}
