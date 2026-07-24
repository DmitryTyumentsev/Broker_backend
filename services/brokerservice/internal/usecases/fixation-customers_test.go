package usecases

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"testing"

	"github.com/google/uuid"
)

type MockNewFixationCustomer struct {
	fixationCustomer entity.FixationCustomer
}

func TestNewFixationCustomer(t *testing.T) {
	mock := &MockNewFixationCustomer{
		fixationCustomer: entity.FixationCustomer{

		}
	}
}
