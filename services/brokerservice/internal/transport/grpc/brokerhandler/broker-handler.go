package brokerhandler

import (
	"Broker_backend/services/brokerservice/internal/usecases/cmd"
	brokerv1 "Broker_backend/shared/pkg/grpc/gen/broker/v1"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface { //зачем нужен интерфейс? вопрос про именно этот кейс - зачем тут добавили интерфейс? что это дает? почему просто не указать service структурой в domain/usecases(не знаю где правильнее)? и вообще - в моей картине мы должны из хендлера брокерсервиса постучаться в usecases. Для этого нам нужно указать ресивером usecases.Service, внутри которого будет cfg, logger, time и тд. Так зачем нам делать интерфейс вместо напрямую структуры? и второй вопрос тут - а методы я какие указываю? юзкейсные? и в юзкейсах будет структура Service? чтобы интерфейс реализовывал. и третье - давай освежим как устроен интерфейс, как выходит что им можно подменять структуры и это будет работать. по памяти - интерфейс хранит методы и ссылки на структуры, верно?
	NewFixationCustomer(ctx context.Context, brokerID cmd.BrokerID, customerID cmd.CustomerID, fixFor cmd.FixFor, fixedBy cmd.FixedBy) error
}
type Handler struct {
	brokerv1.UnimplementedBrokerServiceServer
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{
		service: service,
		logger:  logger,
	}
}
func (h *Handler) CreateFixationCustomer(ctx context.Context, req *brokerv1.ConnectCustomerRequest,
) error {
	if h == nil || h.service == nil { //пишут ли проверки ресивера на nil в самом методе как здесь я написал? разве на уровне этого слоя не вешают валидаторы тоже?
		return status.Error(codes.Unavailable, "broker service is not wired")
	}

	cmdReq := &cmd.FixationCustomerRequest{ //а принято ли использовать подход когда мы между protoDTO/DTO и entity ставим cmdReq? зачем если да?
		CustomerID: req.CustomerId,
		FixFor:     req.FixFor,
		FixedBy:    req.FixedBy,
	}
	err := h.service.NewFixationCustomer(ctx, cmdReq.CustomerID, cmdReq.FixFor, cmdReq.FixedBy)
	if err != nil {
		h.logger.Warn("create customer fixation failed", zap.Error(err)) //почему логируем тут? где принято и как логировать правильно сервисные и инфра ошибки?
		return mapError(err)
	}

	return nil
}
