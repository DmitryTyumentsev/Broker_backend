package brokerhandler

import (
	"Broker_backend/services/brokerservice/internal/usecases/cmd"
	"context"
)

type Handler struct {
	grpc brokerv1.UnimplentedBrokerServiceClient //что писать в этой структуре? верно ли вообще что это хендлер? хендлеры это же вроде про другое?
}

func NewHandler(grpc brokerv1.UnimplentedBrokerServiceClient) *Handler {
	return &Handler{
		grpc: grpc,
	}
}

func (h *Handler) CreateFixationCustomer(ctx context.Context, protoDTO *brokerv1.ConnectCustomerRequest) (*brokerdto.ConnectCustomerResponse, error) {
	cmdReq := cmd.ConnectCustomerRequest{}

	cmdResp, err := usecases.CreateFixationCustomer(ctx, cmdReq)
	if err != nil {
		return nil, domain.MapError(err) //где обычно в больших проектах лежит мапер ошибок сервисного слоя?  что по логированию в сервисном и инфра слоях?
	}
}
