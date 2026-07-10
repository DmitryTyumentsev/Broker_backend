package entity

import "time"

type BrokerID string
type CustomerID string
type ManagerID string

type ConnectCustomerRequest struct {
	CustomerID CustomerID
	BrokerID   BrokerID
	ManagerID  ManagerID
	CreatedAT  *time.Time
}

type ConnectCustomerResponse struct {
	ManagerLastName   string
	ManagerFirstName  string
	ManagerMiddleName string
}
