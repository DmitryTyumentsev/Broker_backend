package interfaces

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
)

type FixationCustomerRepository interface {
	SaveFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID) error
	EditFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID) error
	DeleteFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID) error
	FindManagerByCustomerID(customerID entity.CustomerID) (entity.ManagerID, error) //под вопросом надо ли
}

type FixationCustomerIssue interface {
	CreateFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID) error
	//IsFixationAvailable(customer entity.CustomerID, roleManager string) bool
	IsRaceFixation(customerID entity.CustomerID, managerID entity.ManagerID) bool //надо защититься от гонок чтобы одновременно двое не могли фиксировать кастомера. Так ли вообще делают проверку(выносят в интерфейс отдельный на это метод)?
}
