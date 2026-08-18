package entity

import (
	"time"
)

type User struct {
	ID string
	// AgencyID — агентство сотрудника. NULL у сотрудников застройщика
	// (sales_manager, account_manager): у них агентства нет.
	AgencyID                   *string
	Email                      string
	Role                       UserRole
	PasswordHash               string
	LastName                   string
	FirstName                  string
	MiddleName                 *string
	CreatedAt                  time.Time
	UpdatedAt                  *time.Time
	ReplacedByRefreshTokenHash *string
}
