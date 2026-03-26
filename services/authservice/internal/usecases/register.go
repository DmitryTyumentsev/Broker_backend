package usecases

import (
	"Donate_backend/services/authservice/internal/domain/entity"
	"Donate_backend/services/authservice/internal/pkg/validators"
	"context"
	"fmt"
	"net/mail"

	"github.com/google/uuid"
)

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*TokenPairResponse, error) {
	const op = "usecase.Register"
	if err := validateRegisterInput(req); err != nil {
		return nil, fmt.Errorf("validate error %w", err)
	}

	passHash, err := s.passHasher.Hash(req.Password)
	user := &entity.User{
		ID:       uuid.NewString(),
		Email:    req.Email,
		PassHash: passHash,
		Username: req.Username,
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("couldn’t save user, err: %w, op: %s %w", err, op)
	} //TODO: хорошая ли практика в ошибки писать op всегда когда есть err в методе/функции? есть ли в этом смысл вообще?

	rawRefresh, err := s.refService.New()
	if err != nil {
		return nil, mapError(err)
	}

	hashRefresh, err := s.refService.Hash(rawRefresh)
	if err != nil {
		return nil, mapError(err)
	}

	now := s.clock.Now()
	session := entity.RefreshSession{
		Hash:           hashRefresh,
		UserID:         user.ID,
		DeviceID:       req.DeviceID,
		CreatedAt:      now,
		ExpiresAt:      now.Add(s.config.Business.LifetimeRefreshToken),
		RevokedAt:      nil,
		ReplacedByHash: nil,
	}

	if err := s.refSessRepo.Save(ctx, session); err != nil {
		return nil, mapError(err)
	}

	accessToken, err := s.accessIssier.Issue(user.ID, req.DeviceID)
	if err != nil {
		return nil, mapError(err)
	}

	tokenPair := entity.TokenPairResponse{
		access:       accessToken,
		refresh:      rawRefresh,
		expiresInSec: int64(s.config.Business.LifetimeRefreshToken), //TODO: ты написал что тут указывают access а не refresh, а где тогда указываем refresh?
		// и вообще мы должны же передавать expires_in_sec и access и refresh, так же принято?
		//и браузер клиента будет сохранять это себе в session storage например,
		//так же принято на реальных проектах high load в рф?
	}

	return &TokenPairResponse{
		Access:       tokenPair.access,
		Refresh:      tokenPair.refresh,
		ExpiresInSec: tokenPair.ExpiresInSec,
	}, nil
}

func validateRegisterInput(req *RegisterRequest) (err error) {
	if req.Email == "" { //TODO: можно ли улучшить код или тут лучшее решение все через if писать?
		return ErrEmailRequired
	}
	if !validators.IsEmail(req.Email) {
		return ErrEmailInvalid
	}

	if req.Password == "" {
		return ErrPassRequired
	}
	if !validators.IsStrongPassword(req.Password) {
		return ErrPassEasy
	}

	if req.Username == "" {
		return ErrUsernameRequired
	}
	if !validators.IsUsername(req.Username) {
		return ErrUsernameInvalid
	}
	return nil
}
