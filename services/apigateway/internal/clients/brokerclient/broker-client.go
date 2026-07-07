package brokerclient

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"context"
)

type Client struct {
	grpc brokerv1.BrokerService //вижу что в authclient у меня совсем по другому тут, делается коннект отдельно к grpc и тд. Что мне нужно знать и понимать в этом? как должен быть устроен brokerclient(более важный и точный вопрос)?
}

func (c *Client) CreateFixationCustomer(ctx context.Context, protoDTO *brokerv1.ConnectCustomerRequest) (*brokerdto.ConnectCustomerResponse, error) {
	protoResp, err := c.grpc.CreateFixationCustomer(ctx, protoDTO)
	if err != nil {
		return nil, err // немного запутался - где мапер ошибки? и если на шаг назад - а что нам вообще приходить должно? я так понял что к нам приходит ошибка в виде status.Error и мы должны ее мапить, так? а где это у меня? и какой там контекст передается?
	}
	return protoResp, nil
}
