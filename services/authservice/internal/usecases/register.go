package usecases

import (
	"Donate_backend/services/authservice/internal/domain/entity"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email, Password, Username, DeviceID string
}

type TokenPairResponse struct {
	Access, Refresh string
	ExpiresInSec    int64
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*TokenPairResponse, error) {
	const op = "usecase.Register"
	if err := validateRegister(req); err != nil {
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

func validateRegister(req *RegisterRequest) error {
	//TODO: стандартная либа email называется в /net ?
	if res := email.Validate(req.Email, 2, 255); res == false { //TODO: тут так? return usecase.ErrValidate
		// TODO: в какой момент нужно добавить поле в ошибку? как я понял на сервисном уровне(domain/usecase),
		//мы просто возвращаем errors.New() или fmt.Errorf()? и второй вопрос как верно вернуть ошибку
		//когда в слое ниже есть err? мы должны просто через fmt.Errorf писать в таком случае и err туда передавать?
		//а если у нас под это есть отдельная доменная ошибка? нам же надо и саму err со слоя ниже вернуть. Понял да пример?
		//что нам бд вернула ошибку, и мы в доменном слое ее хотим преобразовать и вернуть. Надо ли вообще err передавать с уровня ниже
		//(думаю что да, но тогда как верно это сделать?)
	}
	if res := validate.PassDefault(req.Password, 8, 72); res == false {
		return errorResponse{}
	}
	if res := validate.Username(req.Username, 2, 50); res == false {
		return errorResponse{}
	}
	return nil
}
