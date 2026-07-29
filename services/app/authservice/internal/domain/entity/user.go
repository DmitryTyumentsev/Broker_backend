package entity

import (
	"time"
)

type User struct {
	ID                         string
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
