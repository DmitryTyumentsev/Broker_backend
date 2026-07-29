package usecases

import (
	"Broker_backend/services/integration/brokerservice/internal/domain"
	"Broker_backend/services/integration/brokerservice/internal/domain/entity"
	"Broker_backend/shared/pkg/helpers"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) NewFixationCustomer(ctx context.Context, cmdFixationCustomer *FixationCustomerRequest) (*entity.FixationCustomer, error) {
	//как лучше сделать - поделить на NewFixation/ExpiredFixation/ConvertFixation/RemovedFixation ? так принято делать в таких и подобных кейсах?
	if s == nil {
		return nil, fmt.Errorf("service not init") //надо ли проверять ресивер на nil? у меня же есть ensureDeps()
	} //у нас есть отдельно apigateway где миддлвар проверяет роль и дает/не дает доступ к ендпоинту. Но у сервисов другой адрес и в целом если есть ендпоинт можно кинуть запрос в обход gateway. По моей логике должны быть корсы которые проверяют адрес и доступный адрес откуда может прийти запрос - наш gateway, а остальные отметать. Так делают на проектах? upd: ще подумал - можно же взять прокси и подменить адрес у корса на наш gateway. Я же верно рассуждаю? как защититься от такого?

	fixationID, err := uuid.NewV6() //был uuid v6 в 2020-2022? он работал также как v4 что unix времени не было и рандомно поэтому дробил и двигал страницы дерева в базе?
	if err != nil {
		return nil, err //тут же ок отдать просто err? моя логика - так как в мапере такой ошибки нет, это будет 500, а текст я смогу тут посмотреть, самого err мне хватит думаю
	}
	fixedAt := s.clock.Now()
	expiresAt := fixedAt.Add(s.cfg.Business.FixationDuration)

	fixationCustomer := entity.FixationCustomer{
		BrokerID:      cmdFixationCustomer.BrokerID,
		PhoneHash:     cmdFixationCustomer.PhoneHash,
		FixFor:        cmdFixationCustomer.FixFor,
		FixedBy:       cmdFixationCustomer.FixedBy,
		FixationIDNew: fixationID,
		FixedAt:       fixedAt,
		ExpiresAt:     expiresAt,
		StatusActive:  entity.StatusActive,
	}

	if err := s.fixations.Insert(ctx, fixationCustomer); err != nil {
		return nil, err
	}

	return &fixationCustomer, nil
}

func (s *Service) UpdateFixation(ctx context.Context, req *brokerdto.FixationRequest, brokerID, userID uuid.UUID) (*entity.FixationCustomer, error) {

	//err := changeStatusToExpired(ctx, req)
	phoneHash := helpers.NormalizeAndHashPhone(req.Phone)

	fixedAt := s.clock.Now()
	expiresAt := fixedAt.Add(s.cfg.Business.FixationDuration)
	fixationIDNew, err := uuid.NewV6()
	if err != nil {
		return nil, err //есть кейсы как тут когда ошибка появляется из-за бага библиотеки. как такие правильно оборачивать? достаточно ли просто err тут?
	}

	f := entity.FixationCustomer{
		BrokerID:      brokerID,
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

	if err := s.fixations.Update(ctx, f); err != nil {
		return nil, err
	}

	resp := &entity.FixationCustomer{
		FixationIDNew: f.FixationIDNew,
		FixedAt:       f.FixedAt,
		ExpiresAt:     f.ExpiresAt,
	}

	return resp, nil
}
