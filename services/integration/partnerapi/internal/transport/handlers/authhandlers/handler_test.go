package authhandlers

import (
	"testing"

	authv1 "Broker_backend/gen/auth/v1"
	"Broker_backend/services/integration/partnerapi/internal/transport/dto/authdto"
)

func TestGRPCToHTTPTokenPair(t *testing.T) {
	tests := []struct {
		name string
		resp tokenPair
	}{
		{
			name: "register",
			resp: &authv1.RegisterResponse{
				AccessToken:  "register-access",
				RefreshToken: "register-refresh",
				ExpiresInSec: 900,
			},
		},
		{
			name: "login",
			resp: &authv1.LoginResponse{
				AccessToken:  "login-access",
				RefreshToken: "login-refresh",
				ExpiresInSec: 900,
			},
		},
		{
			name: "refresh",
			resp: &authv1.RefreshResponse{
				AccessToken:  "refresh-access",
				RefreshToken: "refresh-refresh",
				ExpiresInSec: 900,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grpcToHTTPTokenPair(tt.resp)
			want := authdto.TokenPairResponse{
				Access:       tt.resp.GetAccessToken(),
				Refresh:      tt.resp.GetRefreshToken(),
				ExpiresInSec: 900,
			}

			if got != want {
				t.Fatalf("grpcToHTTPTokenPair() = %+v, want %+v", got, want)
			}
		})
	}
}
