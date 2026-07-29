package usecases

import (
	entity2 "Broker_backend/services/app/authservice/internal/domain/entity"
	validators2 "Broker_backend/services/app/authservice/internal/pkg/validators"
	"context"
	"fmt"
	"strings"

	"Broker_backend/shared/pkg/helpers"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*TokenPairResponse, error) {
	const op = "usecases.Register"

	if err := s.ensureDeps(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := validateRegisterInput(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	passwordHash, err := s.passHasher.Hash(req.RawPassword)
	if err != nil {
		return nil, fmt.Errorf("%s: hash password: %w", op, err)
	}

	now := s.clock.Now()

	user := entity2.User{
		ID:           uuid.NewString(),
		Email:        helpers.NormalizeEmail(req.Email),
		Role:         entity2.RoleBrokerTeamMember,
		PasswordHash: passwordHash,
		LastName:     helpers.NormalizeText(req.LastName),
		FirstName:    helpers.NormalizeText(req.FirstName),
		MiddleName:   helpers.NormalizeOptionText(req.MiddleName),
		CreatedAt:    now,
	}

	if err := s.users.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("%s: save user: %w", op, err)
	}

	tokenPair, err := s.createTokenPair(ctx, user.ID, req.DeviceID, string(user.Role), now)
	if err != nil {
		return nil, fmt.Errorf("%s: create token pair: %w", op, err)
	}

	s.logger.Info("user registered", zap.String("user_id", user.ID))

	return tokenPair, nil
}

func validateRegisterInput(req *RegisterRequest) error {
	if req == nil {
		return ErrEmailRequired
	}

	req.Email = helpers.NormalizeEmail(req.Email)
	req.FirstName = helpers.NormalizeText(req.FirstName)
	req.LastName = helpers.NormalizeText(req.LastName)
	req.DeviceID = strings.TrimSpace(req.DeviceID)

	if req.Email == "" {
		return ErrEmailRequired
	}

	if !validators2.IsValidEmail(req.Email) {
		return ErrEmailInvalid
	}

	if req.RawPassword == "" {
		return ErrPasswordRequired
	}

	if !validators2.IsStrongPassword(req.RawPassword) {
		return ErrPasswordWeak
	}

	if req.DeviceID == "" {
		return ErrDeviceIDRequired
	}

	if req.FirstName == "" {
		return ErrFirstNameRequired
	}

	if req.LastName == "" {
		return ErrLastNameRequired
	}

	return nil
}
