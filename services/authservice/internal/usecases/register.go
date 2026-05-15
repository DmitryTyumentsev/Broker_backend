package usecases

import (
	"Donate_backend/shared/pkg/helpers"
	"context"
	"fmt"
	"strings"

	"Donate_backend/services/authservice/internal/domain/entity"
	"Donate_backend/services/authservice/internal/pkg/validators"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	roleSuperadmin       = "superAdmin"       //технический админ всей платформы
	roleDeveloperAdmin   = "developerAdmin"   //админ застройщика
	roleAccountManager   = "accountManager"   //менеджер застройщика по работе с агентствами
	roleSalesManager     = "salesManager"     //менеджер продаж застройщика по конкретным сделкам
	roleAgencyOwner      = "agencyOwner"      //руководитель агентства недвижимости
	roleBrokerTeamLeader = "brokerTeamLeader" //руководитель группы брокеров внутри агентства
	roleBrokerTeamMember = "brokerTeamMember" //брокер / агент
)

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*TokenPairResponse, error) { //в методе вижу что данные передаю где-то по ссылке, где-то без ссылки. Данные между слоями вообще как верно передавать? у меня не будут копии в каждом методе создаваться если без ссылки буду их передавать? у меня тут пробел как будто
	const op = "usecases.Register"

	if s == nil {
		return nil, fmt.Errorf("%s: service is nil", op)
	}

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

	user := entity.User{
		ID:           uuid.NewString(),
		Email:        strings.TrimSpace(req.Email),
		Role:         roleBrokerTeamMember,
		PasswordHash: passwordHash,
		LastName:     helpers.NormalizeString(req.LastName),
		FirstName:    helpers.NormalizeString(req.FirstName),
		MiddleName:   helpers.NormalizeString(req.MiddleName), //горит красным из-за *string. Как на продовых проектах делают? мне helper менять или прям тут поправить как-то?
		CreatedAt:    now,
	}

	if err := s.users.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("%s: save user: %w", op, err)
	}

	rawRefreshToken, err := s.refreshToken.New()
	if err != nil {
		return nil, fmt.Errorf("%s: create refresh token: %w", op, err)
	}

	refreshHash := s.refreshToken.Hash(rawRefreshToken)

	session := entity.RefreshSession{
		RefreshTokenHash: refreshHash,
		DeviceID:         req.DeviceID,
		CreatedAt:        now,
		ExpiresAt:        now.Add(s.config.Business.LifetimeRefreshToken),
	}

	if err := s.sessions.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("%s: save refresh session: %w", op, err)
	}

	accessToken, err := s.accessIssuer.Issue(user.ID, req.DeviceID, user.Role, now)
	if err != nil {
		return nil, fmt.Errorf("%s: issue access token: %w", op, err)
	}

	if s.logger != nil {
		s.logger.Info("user registered", zap.String("user_id", user.ID))
	}

	return &TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresInSec: int64(s.config.Business.LifetimeAccessToken.Seconds()),
	}, nil
}

func validateRegisterInput(req *RegisterRequest) error {
	if req == nil {
		return ErrEmailRequired
	}

	arrayReq := make([]string, ) //как выделить объем под *struct? была идея попробовать из req вытащить все поля и проверить на "" их в цикле for range. Как тебе? как сделать?
	for k, v := range

	if req.Email == "" {
		return ErrEmailRequired
	}
	if req.RawPassword == "" {
		return ErrPasswordRequired
	}
	if req.DeviceID == "" {
		return ErrDeviceIDRequired
	}

	req.Email = helpers.NormalizeString(req.Email)

	if !validators.IsValidEmail(req.Email) {
		return ErrEmailInvalid
	}

	if !validators.IsStrongPassword(req.RawPassword) {
		return ErrPasswordWeak
	}

	return nil
}
