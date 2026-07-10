package interfaces

import (
	"Broker_backend/services/brokerservice/internal/domain"
	"time"
)

type Clock interface {
	Now() time.Time
}

type ManagerRepository interface {
	RoleByManagerID(managerID domain.ManagerID) (string, error)
	FullNameByManagerID(managerID domain.ManagerID) (string, error)
}
