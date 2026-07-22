package usecases

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"Broker_backend/services/brokerservice/internal/usecases/cmd"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) NewFixationCustomer(ctx context.Context, brokerID cmd.BrokerID, customerID cmd.CustomerID, fixFor cmd.FixFor, fixedBy cmd.FixedBy) (*cmd.FixationCustomerResponse, error) {
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

	reqEntity := &entity.FixationCustomer{
		BrokerID:   entity.BrokerID(brokerID),
		CustomerID: entity.CustomerID(customerID),
		FixFor:     entity.FixFor(fixFor),
		FixedBy:    entity.FixedBy(fixedBy),
	}

	if err := s.fixations.Insert(ctx, fixationID, fixedAt, expiresAt, entity.StatusActive, reqEntity.BrokerID, reqEntity.FixedBy, reqEntity.FixFor, reqEntity.CustomerID); err != nil {
		return nil, err
	}

	respEntity := &entity.FixationCustomer{ //юзкейс должен отдавать в хендлер entity или cmdResp? давай еще раз тут, верно ли чтобы наш хендлер получил entity, не понимаю логику до конца между хендлером нашего сервиса и его юзкейсом
		FixationID: fixationID,
		FixedAt:    entity.FixedAt(fixedAt),
		ExpiresAt:  entity.ExpiresAt(expiresAt),
		Status:     entity.StatusActive,
	}

	return respEntity, nil
}
