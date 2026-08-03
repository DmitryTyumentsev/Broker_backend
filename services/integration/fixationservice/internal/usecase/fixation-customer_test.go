package usecase

import (
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
)

type MockNewFixationCustomer struct {
	fixationCustomer entity.Fixation
}

//
//func TestNewFixationCustomer(t *testing.T) {
//	mock := &MockNewFixationCustomer{
//		fixationCustomer: entity.Fixation{
//
//		}
//	}
//}
