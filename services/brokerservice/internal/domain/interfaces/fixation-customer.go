package interfaces

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"time"
)

type FixationCustomerRepository interface {
	SaveFixationCustomer(uuid string, expiresAT, fixedAt *time.Time, status string, brokerID entity.BrokerID, fixedBy entity.FixedBy, managerID entity.ManagerID, customerID entity.CustomerID) error
	UpdateFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID) error
	FreeFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID) error
	SelectStatusByCustomerID(customerID entity.CustomerID) (string, error) //под вопросом надо ли
}

type FixationCustomerService interface { //принято ли делить интерфейсы в сервисном уровне отдельно на Issue, Check и тд или просто когда фичю пилишь - создаешь интерфейс под сервис и под репо?
	CreateFixationCustomer(brokerID entity.BrokerID, fixedBy entity.FixedBy, managerID entity.ManagerID, customerID entity.CustomerID) error
	CheckStatusCustomer(customerID entity.CustomerID) (string, error)
}
