package entity

import "time"

type User struct {
	ID           string
	Email        string
	Role         string
	PasswordHash string
	LastName     string
	FirstName    string
	MiddleName   *string
	CreatedAt    time.Time //опять забыл правило - если поле обязательное - оставляю так. если необязательное - делаю *Type? нужно это чтобы сделать логику где-то(кстати где именно?), что если строка == nil, то что-то сделай(кстати что?). Вообщем нужен пример чтобы я лучше понимал если вообще щас верно понял зачем это
}
