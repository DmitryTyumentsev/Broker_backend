package brokerdto

type ConnectCustomerRequest struct {
	CustomerID string `json:"customer_id" validate:"required,uuid4"` //мешают ли для мапингу приватные типы, если да - почему?
	BrokerID   string `json:"broker_id" validate:"required"`
	ManagerID  string `json:"manager_id" validate:"omitempty"`
}

type ConnectCustomerResponse struct {
	ManagerLastName   string `json:"manager_last_name"`
	ManagerFirstName  string `json:"manager_first_name"`
	ManagerMiddleName string `json:"manager_middle_name"`
}
