package usecases

import (
	"Donate_backend/services/authservice/internal/domain"
	"Donate_backend/services/authservice/internal/domain/entity"
	"Donate_backend/services/authservice/internal/infra/security/jwt"
	"context"
	"fmt"
)

func (s *Service) Auth(ctx context.Context, req *AuthRequest) (resp *TokenPairResponse, err error) {
	const op = "usecases.Auth"
	s.logger.Debug(fmt.Sprintf("start %s", op)) //надо ли на zap вешать defer close в конце каждого метода где вызывался(или как он правильно но с этим смыслом)

	user, err := s.findUser(ctx, req.Email, req.RawPassword)
	if err != nil {
		return nil, err
	}

	resp, err = s.createSession(ctx, user.ID, user.Role, req.DeviceID)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Service) findUser(ctx context.Context, email, rawPass string) (*entity.User, error) { //по названию верно ли писать rawPass тут а в миграции password_hasher? какое название вообще будет правильным по синтаксису имени?
	const op = "usecases.checkCredentials"
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Warn(fmt.Sprintf("%s: %s", op, err.Error()))
		return nil, domain.ErrNotFound
	}

	if s.passHasher.Verify(user.PasswordHash, rawPass) == false {
		s.logger.Warn(fmt.Sprintf("%s: %s", op, domain.ErrPasswordWrong))
		return nil, domain.ErrPasswordWrong
	}

	return &user, nil
}

func (s *Service) createSession(ctx context.Context, userID string, role entity.UserRole, deviceID string) (*TokenPairResponse, error) {
	const op = "usecases.createSession"
	rawRefresh, err := s.refreshToken.New()
	if err != nil {
		return nil, err
	}
	hashRefresh := s.refreshToken.Hash(rawRefresh)

	refreshSession := jwt.RefreshSession{
		RefreshTokenHash: hashRefresh,
		DeviceID:         deviceID,
		CreatedAt:        s.clock.Now(),
		ExpiresAt:        s.clock.Now().Add(s.config.Business.LifetimeRefreshToken),
	}
	if err = s.sessions.Save(ctx, refreshSession); err != nil {
		s.logger.Error(fmt.Sprintf("%s: %s", op, err.Error()))
		return nil, domain.ErrSessionNotSaved
	}
	accessToken, err := s.accessIssuer.Issue(userID, deviceID, role, s.clock.Now())
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s: %s", op, err.Error()))
		return nil, domain.ErrAccessTokenNotCreated //верно и хорошая ли практика указывать так подробно на этом слое ошибки, если потом они будут мапиться в badRequest на уровне клиента?
	}

	resp := &TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshSession.RefreshTokenHash,
		ExpiresInSec: s.clock.Now().Add(s.config.Business.LifetimeAccessToken).Unix(), //какой правильно возвращать тип? int64 или time.Second?
	}

	return resp, nil
}
