package brokerclient

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"context"
)

type Client struct {
	grpc brokerv1.UnimplentedBrokerServiceClient //вижу что в authclient у меня совсем по другому тут, делается коннект отдельно к grpc и тд. Что мне нужно знать и понимать в этом?
}

func (c *Client) ConnectCustomer(ctx context.Context, dto *brokerdto.ConnectCustomerRequest) (*brokerdto.ConnectCustomerResponse, error) {
	resp, err := c.grpc.ConnectCustomer(ctx, dto) //где как правило dto преобразуют в cmd? на клиенте или на транспорте куда отправляем из apigateway?
	if err != nil {
		return nil, err // немного запутался - где мапер ошибки? и если на шаг назад - а что нам вообще приходить должно? я так понял что к нам приходит ошибка в виде status.Error и мы должны ее мапить, так? а где это у меня? и какой там контекст передается?
	}
}
