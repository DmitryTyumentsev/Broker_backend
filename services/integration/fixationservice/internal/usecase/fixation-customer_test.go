package usecase

import (
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
)

type MockNewFixationCustomer struct {
	fixationCustomer entity.FixationCustomer
}

//
//func TestNewFixationCustomer(t *testing.T) {
//	mock := &MockNewFixationCustomer{
//		fixationCustomer: entity.FixationCustomer{
//
//		}
//	}
//}
