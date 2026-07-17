package interfaces

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"time"
)

type FixationCustomerRepository interface {
	SaveActiveFixationCustomer(uuid string, expiresAT, fixedAt *time.Time, statusActive string, brokerID entity.BrokerID, fixedBy entity.FixedBy, fixFor entity.FixFor, customerID entity.CustomerID) error
	UpdateFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID) error
	FreeFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID) error
	SelectStatusByCustomerID(customerID entity.CustomerID) (string, error) //под вопросом надо ли
}

type FixationCustomerService interface { //принято ли делить интерфейсы в сервисном уровне отдельно на Issue, Check и тд или просто когда фичю пилишь - создаешь интерфейс под сервис и под репо?
	CreateActiveFixationCustomer(brokerID entity.BrokerID, fixedBy entity.FixedBy, fixFor entity.FixFor, customerID entity.CustomerID) error
	CheckStatusCustomer(customerID entity.CustomerID) (string, error)
}
