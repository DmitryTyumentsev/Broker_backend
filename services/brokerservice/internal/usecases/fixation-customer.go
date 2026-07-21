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
	now := s.clock.Now()
	expiresAt := now.Add(s.cfg.Business.FixationDuration)

	reqEntity := &entity.FixationCustomerRequest{
		BrokerID:   entity.BrokerID(brokerID), // вот и ошибка с типами. Это в продолжение вопроса из пакета cmd про типы - где уместно вешать а где нет. Оставлю ошибку так чтобы ты мог увидеть кейс и объяснить. с одной стороны надо не спутать случайно переменные с одним типом чтобы подставить в базу корректно, с другой такая ошибка. верно понял что правильно тут создать только под entity типы потому что идем в постгрю и там важен порядок а в cmd без разницы так как конвертируем дальше?
		CustomerID: entity.CustomerID(customerID),
		FixFor:     entity.FixFor(fixFor),
		FixedBy:    entity.FixedBy(fixedBy),
	}

	if err := s.fixations.Insert(ctx, fixationID, now, expiresAt, entity.StatusActive, reqEntity.BrokerID, reqEntity.FixedBy, reqEntity.FixFor, reqEntity.CustomerID); err != nil {
		return nil, err
	}

	respEntity := &entity.FixationCustomerResponse{ //юзкейс должен отдавать в хендлер entity или cmdResp?
		FixationID: fixationID,
		FixedAt:    now,
		ExpiresAt:  expiresAt,
		Status:     entity.StatusActive,
	}

	return respEntity, nil
}
