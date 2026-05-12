package jwt

import (
	"errors"
	"time"
)

type AccessTokenIssuer struct{}

func NewAccessTokenIssuer() *AccessTokenIssuer {
	return &AccessTokenIssuer{}
}

func (i *AccessTokenIssuer) Issue(
	userID string,
	deviceID string,
	role string,
	now time.Time,
) (string, error) {
	return "", errors.New("access token issuer is not implemented yet")
}
