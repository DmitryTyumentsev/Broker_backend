package usecases

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"context"
	"fmt"
)

func (s *Service) CreateFixationCustomer(ctx context.Context, brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID) error {
	if s == nil {
		return fmt.Errorf("service not init") //надо ли проверять ресивер на nil? у меня же есть ensureDeps()
	} //у нас есть отдельно apigateway где миддлвар проверяет роль и дает/не дает доступ к ендпоинту. Но у сервисов другой адрес и в целом если есть ендпоинт можно кинуть запрос в обход gateway. По моей логике должны быть корсы которые проверяют адрес и доступный адрес откуда может прийти запрос - наш gateway, а остальные отметать. Так делают на проектах? upd: ще подумал - можно же взять прокси и подменить адрес у корса на наш gateway. Я же верно рассуждаю? как защититься от такого?
	now := s.clock.Now()

	if managerID == "" {
		return s.pg.SaveFixationCustomer(ctx, brokerID, managerID, customerID, &now) //хочу запустить в горутине, но не понимаю на логике где писать горутины и по синтаксису увидел что неправильно пишу(стер что писал, закомменченное ниже это то как писал)
		//опять забыл как в таком кейсе передать оригинал чтобы копия не создавалась переменной
	}
	//if managerID == "" {
	//		err := go func() error {
	//			return s.pg.SaveFixationCustomer(ctx, brokerID, managerID, customerID, &now) //хочу запустить в горутине, но вижу что неправильно пишу
	//		} ()
	//	}

	return s.pg.EditFixationCustomer(ctx, brokerID, managerID, customerID, &now) //надо запихнуть все в транзакцию у бд,
}
