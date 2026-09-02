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
	status, err := s.fixations.StatusByProjectID(ctx, req.ProjectID)
	if err != nil {
		if status == "" {
			return nil, domain.ErrProjectNotExist
		}
		return nil, err
	}
	if status == entity.StatusProjectArchived {
		return nil, domain.ErrProjectArchived
	}
	b, err := s.fixations.IsUserIDInAgencyID(ctx, req.AgencyID, req.FixFor)
	if b == false {
		return nil, domain.ErrEmployeeNotInAgency
	}
	if err != nil {
		return nil, err
	}

	phoneHash, err := s.HashPhone(req.Phone)
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

func (s *Service) HashPhone(phone string) (string, error) {
	normalizePhone := helpers.NormalizePhoneNumber(phone)

	return security.SignHS256AndEncodeBase64URL(s.cfg.Business.HashSecret, normalizePhone)
}
