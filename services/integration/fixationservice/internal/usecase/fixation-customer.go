package usecase

import (
	"Broker_backend/services/integration/fixationservice/internal/domain"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"Broker_backend/services/integration/fixationservice/internal/infra/security"
	"Broker_backend/shared/pkg/helpers"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) NewFixation(ctx context.Context, req *FixationRequest) (*entity.Fixation, error) {
	b, err := s.fixations.IsExistsProjectID(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if b == false {
		return nil, err
	}
	b, err = s.fixations.IsUserIDInAgencyID(ctx, req.AgencyID, req.FixFor)
	if err != nil {
		return nil, err
	}
	if b == false {
		return nil, err
	}

	phone := helpers.NormalizePhoneNumber(req.Phone)
	phoneHash, err := security.SignHS256AndEncodeBase64URL(s.cfg, phone)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	expiresAt := now.Add(s.cfg.Business.FixationDuration)
	fixationID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	newFixation := entity.Fixation{
		AgencyID:   req.AgencyID,
		PhoneHash:  phoneHash,
		FixFor:     req.FixFor,
		FixBy:      req.FixBy,
		FixationID: fixationID,
		FixedAt:    now,
		ExpiresAt:  expiresAt,
		Status:     entity.StatusActive,
		ProjectID:  req.ProjectID,
	}

	err = s.tx.Do(ctx, func(txCtx context.Context) error {
		fixationFromDB, err := s.fixations.FixationCurrent(txCtx, phoneHash, req.ProjectID)
		if err != nil {
			return err
		}
		switch fixationFromDB.Status {
		case entity.StatusConverted:
			return domain.ErrFixationAlreadyExist
		case entity.StatusActive:
			if fixationFromDB.ExpiresAt.Before(now) == true || fixationFromDB.ExpiresAt.Equal(now) == true {
				err = s.fixations.UpdateFixationStatusExpired(txCtx, entity.StatusExpired, fixationFromDB.FixationID)
				if err != nil {
					return err
				}
				err = s.insertFixationList(txCtx, newFixation)
				if err != nil {
					return err
				}
			}
			if fixationFromDB.ExpiresAt.After(now) == true {
				return domain.ErrFixationAlreadyExist
			}
			return nil

		case entity.StatusExpired, entity.StatusRemoved, entity.StatusNoRows:
			return s.insertFixationList(txCtx, newFixation)
		default:
			return nil
		}
	})
	if err != nil {
		return nil, fmt.Errorf("transaction error: %w", err)
	}

	return &newFixation, nil
}

func (s *Service) insertFixationList(txCtx context.Context, f entity.Fixation) error {
	err := s.fixations.InsertNewFixation(txCtx, f)
	if err != nil {
		return err
	}
	err = s.fixations.InsertAudit(txCtx, f)
	if err != nil {
		return err
	}
	err = s.fixations.InsertOutbox(txCtx, f)
	if err != nil {
		return err
	}
	return nil
}
