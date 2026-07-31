package usecase

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/http/dto/fixationdto"
	"Broker_backend/services/integration/fixationservice/internal/domain"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"Broker_backend/shared/pkg/helpers"
	"context"
	"encoding/base64"

	"github.com/google/uuid"
)

func (s *Service) NewFixation(ctx context.Context, req *fixationdto.FixationRequest, agencyID, userID uuid.UUID) (*entity.FixationCustomer, error) {
	//в базе телефон хранится как хэш. если хэшировать - каждый раз хэш будет разный. как мне проверить есть такой номер или нет?

	if isExistsProjectID(ctx, req.ProjectID) == false {
		return nil, domain.ErrBadRequest
	}

	phone := helpers.NormalizePhoneNumber(req.Phone)
	phoneHash := base64.StdEncoding.EncodeToString([]byte(phone))
	fixation, err := s.fixations.FindActiveFixation(ctx, phoneHash, req.ProjectID)

	fixationID, err := uuid.NewV6() //был uuid v6 в 2020? он работал также как v4 что unix времени не было и рандомно поэтому дробил и двигал страницы дерева в базе?
	if err != nil {
		return nil, err //тут же ок отдать просто err? моя логика - так как в мапере такой ошибки нет, это будет 500, а текст я смогу тут посмотреть, самого err мне хватит думаю
	}
	fixedAt := s.clock.Now()
	expiresAt := fixedAt.Add(s.cfg.Business.FixationDuration)

	f := entity.FixationCustomer{
		AgencyID:      agencyID,
		PhoneHash:     phoneHash,
		FixFor:        req.FixFor,
		FixedBy:       userID,
		FixationIDNew: fixationIDNew,
		FixationIDOld: req.FixationIDOld,
		FixedAt:       fixedAt,
		ExpiresAt:     expiresAt,
		StatusActive:  entity.StatusActive,
		StatusExpired: entity.StatusExpired,
		ProjectID:     req.ProjectID,
	}

	if err := s.fixations.Insert(ctx, f); err != nil {
		return nil, err
	}

	return &f, nil
}

func findProjectID(req *fixationdto.FixationRequest) (*fixationdto.FixationRequest, error) {

	checkProjectID
}
