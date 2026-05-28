package grpchandler

import (
	"context"
	"errors"
	"fmt"

	"Broker_backend/services/authservice/internal/usecases"
	authv1 "Broker_backend/shared/pkg/grpc/gen/auth/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer

	service *usecases.Service
	logger  *zap.Logger
}

func NewHandler(service *usecases.Service) *Handler {
	return &Handler{
		service: service,
	}
} //вопрос по структуре сервиса - тут сейчас есть только transport, транспорт отвечает за принятие от других сервисов. Но в текущей реализации он еще и дает ответ, то есть ведет себя и как клиент который дает ответ. Немного путаница, верно ли все у меня сейчас? или правильно смотреть по первому вызову - принимает или отдает данные и в зависимости от этого делить? путаница в том что сейчас транспорт одновременно и принимает и отдает данные туда же откуда принял

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.TokenPairResponse, error) {
	if h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired yet")
	}

	resp, err := h.service.Register(ctx, &usecases.RegisterRequest{
		Email:       req.Email,
		RawPassword: req.Password,
		LastName:    req.LastName,
		FirstName:   req.FirstName,
		MiddleName:  req.MiddleName,
		DeviceID:    req.DeviceId,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.TokenPairResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresInSec: resp.ExpiresInSec,
	}, nil
}

func (h *Handler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.TokenPairResponse, error) {
	const op = "transport.grpchandler.Login"
	if h.service == nil || h == nil { //почему горит желтым h==nil? верно задал условие по синтаксису?
		h.logger.Error(fmt.Sprintf("%s, service or handler is nil", op)) //во-первых верно ли к каждой ошибке дописывать логирование? когда логировать ошибки, все ли? и вообще зачем нужен логер когда ошибки же и так будут в терминале показываться? во-вторых, верно ли для передачи op использовать fmt.Sprintf? хорошее ли это решение? смущает что в терминал будет писаться, но в этом же и суть логера? и в третьих, хорошо ли из сервиса тянуть логгер сюда, не лучше ли его отдельно сюда передавать? upd: так и переделал, напиши верно ли думаю
		return nil, errors.New("service or handler is nil")
	}

	authReq := &usecases.LoginRequest{ //ок ли такое название давать или лучше указывать слой например назвать переменную usecasesReq?
		Email:       req.Email,
		RawPassword: req.Password,
		DeviceID:    req.DeviceId,
	}

	tokenPair, err := h.service.Login(ctx, authReq)
	if err != nil {
		h.logger.Error(fmt.Sprintf("%s, %s", op, err.Error()))
		return nil, mapError(err)
	}

	if tokenPair == nil {
		h.logger.Error(fmt.Sprintf("%s, %s", op, "empty response"))
		return nil, status.Error(codes.InvalidArgument, "empty response")
	}
	resp := &authv1.TokenPairResponse{
		AccessToken:  tokenPair.AccessToken, //не понимаю как прописать правильно чтобы не было проблем с nil. я должен возвращать структуру по указателю, но как поправить проблему с возможным nil не понимаю. Хочу понять именно как на продовых проектах пишут, ведь вроде все верно - передаем значение по ссылке между слоями. upd: попробовал сделать с проверкой на nil, напиши хорошая это практика или нет отдельно
		RefreshToken: tokenPair.RefreshToken,
		ExpiresInSec: tokenPair.ExpiresInSec,
	}

	return resp, nil
}
