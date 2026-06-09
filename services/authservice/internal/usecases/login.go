package usecases

import (
	"Broker_backend/services/authservice/internal/domain"
	"Broker_backend/services/authservice/internal/pkg/validators"
	"Broker_backend/shared/pkg/helpers"
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*TokenPairResponse, error) {
	const op = "usecases.Login"

	if err := s.ensureDeps(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := validateLoginInput(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrUnauthenticated)
		}

		return nil, fmt.Errorf("%s: find user by email: %w", op, err)
	}

	if !s.passHasher.Verify(user.PasswordHash, req.RawPassword) {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrUnauthenticated)
	}

	now := s.clock.Now()

	tokenPair, err := s.createTokenPair(ctx, user.ID, req.DeviceID, string(user.Role), now)
	if err != nil {
		return nil, fmt.Errorf("%s: create token pair: %w", op, err)
	}

	s.logger.Info("user logged in", zap.String("user_id", user.ID))

	//var wg sync.WaitGroup
	//wg.Add(1)
	//
	//chTokenPair := make(chan *TokenPairResponse)
	//
	//go func(ctx context.Context, userID, deviceID string, userRole entity.UserRole, now time.Time) chan *TokenPairResponse {
	//	defer wg.Done() //напомни defer после return или перед
	//
	//	tokenPair, err := s.createTokenPair(ctx, userID, deviceID, string(user.Role), now)
	//	if err != nil {
	//		return fmt.Errorf("%s: create token pair: %w", op, err)
	//	}
	//
	//	s.logger.Info("user logged in", zap.String("user_id", user.ID))
	//	chTokenPair <- tokenPair
	//	return chTokenPair
	//} (ctx, user.ID, req.DeviceID, user.Role, now)
	//
	//wg.Wait()
	//
	//tokenPair := <- chTokenPair

	return tokenPair, nil
}

func validateLoginInput(req *LoginRequest) error {
	if req == nil {
		return ErrEmailRequired
	}

	req.Email = helpers.NormalizeEmail(req.Email)
	req.DeviceID = strings.TrimSpace(req.DeviceID)

	if req.Email == "" {
		return ErrEmailRequired
	}

	if !validators.IsValidEmail(req.Email) {
		return ErrEmailInvalid
	}

	if req.RawPassword == "" {
		return ErrPasswordRequired
	}

	if req.DeviceID == "" {
		return ErrDeviceIDRequired
	}

	return nil
}
