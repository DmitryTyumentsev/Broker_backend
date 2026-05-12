package entity

import "time"

type User struct {
	ID        string
	Email     string
	Username  string
	PassHash  string
	Role      string
	CreatedAt time.Time
}
