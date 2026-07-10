package usecases

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"testing"

	"github.com/google/uuid"
)

const (
	countGorutines = 10
)

type MockFixationCustomer struct {
	CustomerID entity.CustomerID
	BrokerID   entity.BrokerID
	ManagerID  entity.ManagerID
}

func TestCreateFixationCustomer(t *testing.T) { //начал писать тест и понял что не понял зачем и на что я его пишу. На что писать тест? как?
	mock := &MockFixationCustomer{
		CustomerID: entity.CustomerID(uuid.NewString()),
		BrokerID:   entity.BrokerID(uuid.NewString()),
		ManagerID:  "",
	}

	for i := 0; i < countGorutines; i++ {
		go func() {

		}()
	}
}
