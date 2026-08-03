package usecase

import (
	fixationv1 "Broker_backend/gen/fixation/v1"
	"Broker_backend/services/integration/fixationservice/internal/domain"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"Broker_backend/shared/pkg/helpers"
	"context"
	"encoding/base64"

	"github.com/google/uuid"
)

func (s *Service) NewFixation(ctx context.Context, req *fixationv1.NewFixationRequest) (*entity.Fixation, error) {
	//в базе телефон хранится как хэш. если хэшировать - каждый раз хэш будет разный. как мне проверить есть такой номер или нет?

	if s.fixations.IsExistsProjectID(ctx, req.ProjectId) == false {
		return nil, domain.ErrBadRequest
	}
	if s.fixations.IsUserIDInAgencyID(ctx, req.AgencyId, req.FixFor) == false {
		return nil, domain.ErrForbidden
	}

	phone := helpers.NormalizePhoneNumber(req.Phone)
	phoneHash := base64.StdEncoding.EncodeToString([]byte(phone))
	fixedAt := s.clock.Now()
	expiresAt := fixedAt.Add(s.cfg.Business.FixationDuration)
	fixationID, err := uuid.NewV6() //был uuid v6 в 2020? он работал также как v4 что unix времени не было и рандомно поэтому дробил и двигал страницы дерева в базе?
	if err != nil {
		return nil, err //тут же ок отдать просто err? моя логика - так как в мапере такой ошибки нет, это будет 500, а текст я смогу тут посмотреть, самого err мне хватит думаю
	}

	f := entity.Fixation{
		AgencyID:   req.AgencyId,
		PhoneHash:  phoneHash,
		FixFor:     req.FixFor,
		FixedBy:    req.FixFor,
		FixationID: fixationID,
		FixedAt:    fixedAt,
		ExpiresAt:  expiresAt,
		Status:     entity.StatusActive,
		ProjectID:  req.ProjectId,
	}

	err = s.tx.Do(ctx, func(txCtx context.Context) error {
		status, err := s.fixations.FixationCurrentStatus(txCtx, phoneHash, req.ProjectId)
		switch status {
		case entity.StatusActive, entity.StatusConverted:
			return domain.ErrFixationAlreadyExist
		case entity.StatusExpired, entity.StatusRemoved:
			err = s.fixations.InsertFixation(txCtx, f)
			if err != nil {
				return domain.ErrBadRequest
			}
			err = s.fixations.InsertAudit(txCtx, f)
			if err != nil {
				return domain.ErrBadRequest
			}
			err = s.fixations.InsertOutbox(txCtx, f)
			if err != nil {
				return domain.ErrBadRequest
			}
		default:
			return domain.ErrGeneral
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &f, nil
}
