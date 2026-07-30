package brokerhandlers

import (
	"Broker_backend/services/apigateway/internal/clients/brokerclient"
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"Broker_backend/services/apigateway/internal/transport/http/httperr"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"
	"context"

	validate "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BrokerHandler struct {
	logger     *zap.Logger
	httpClient *brokerclient.HTTPClient
	validator  *validate.Validate
	fixation   FixationClient
}

func NewBrokerHandler(logger *zap.Logger, validator *validate.Validate) *BrokerHandler {
	return &BrokerHandler{
		logger:    logger,
		validator: validator,
	}
}

type FixationClient interface {
	NewFixation(ctx context.Context, req *brokerdto.FixationRequest, agencyID, userID uuid.UUID) (*brokerdto.FixationResponse, error)
}

func (h *BrokerHandler) NewFixation(c *fiber.Ctx) error {
	dtoReq, ok := middleware.ValidatedBody[brokerdto.FixationRequest](c)
	if !ok {
		return httperr.WriteBadRequest(c, "invalid request")
	}
	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return httperr.WriteBadRequest(c, "invalid request")
	}

	ctx := c.UserContext() //зачем перекладываем из fiber.Ctx в context.Context? разве нельзя fiber.Ctx дальше передавать?

	dtoResp, err := h.fixation.NewFixation(ctx, &dtoReq, principal.AgencyID, principal.UserID)
	if err != nil {
		return httperr.WriteHTTPError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dtoResp)
}

//
//func (h *BrokerHandler) Validate() error {
//	switch {
//	case h == nil:
//		return errors.New("brokerservice handler is nil")
//	case h.grpcClient == nil:
//		return errors.New("brokerservice grpcClient is required")
//	default:
//		return h.grpcClient.Validate()
//	}
//}
//
//func (h *BrokerHandler) NewFixation(c *fiber.Ctx) error { //Верно понял что согласно моему миддлвару аксесс лог, каждый(вообще каждый) вызов всех методов из цепочки очень подробно записывается и трейсится ещё плюсом?
//	bodyDTO, ok := middleware.ValidatedBody[brokerdto.FixationRequest](c)
//	if ok == false {
//		h.logger.Error("middleware.ValidatedBody error: type dto didn't match with c.Locals(validatedBodyKey)")
//		return c.JSON(httperr.WriteBadRequest(c, "invalid request"))
//	}
//	principal, ok := middleware.CurrentPrincipal(c) //почитал код, вроде у меня уже есть принципал через middleware.Auth. А как его вытащить, как я сейчас написал? зачем тогда в c клали?
//	if !ok {
//	}
//	fixedBy := principal.UserID
//	protoDTO := &brokerv1.NewFixationCustomerRequest{ //правильно ли вообще передавать managerID, brokerID или можно как-то проще, например просто из принципала вытаскивать? как делают на больших проектах? второй вопрос - как проверять что мне не подставляют чужие данные в запросе? делают ли это в хендлере или миддлварах, есть ли вообще у меня это?
//		BrokerId:   bodyDTO.AgencyID,
//		CustomerId: bodyDTO.Phone,
//		FixedBy:    fixedBy,
//		FixFor:     bodyDTO.FixFor,
//	}
//
//	ctx := c.UserContext() //что у нас будет внутри ctx? у нас был *fiber.Ctx в котором конфиги, данные. А в context.Context и то и то уйдет? что в нем будет?
//	protoResp, err := h.grpcClient.NewFixation(ctx, protoDTO)
//	if err != nil {
//		middleware.AuditLog(
//			c,
//			h.logger,
//			"create fixation customer is failed",
//			zap.Error(err), //стоит ли так писать в аудит логах и почему?
//			zap.String("broker_id", bodyDTO.AgencyID),
//			zap.String("customer_id", bodyDTO.Phone),
//			zap.String("fix_for", bodyDTO.FixFor),
//		)
//		h.logger.Error("grpcClient.NewFixation error", zap.Error(err)) //какой формат у ошибок в больших проектах в таких ситуациях пишут? что тут писать и зачем если у нас такой подробный access logger? как на больших проектах принято? и второй вопрос - как тут правильнее писать по уровню ошибки - это warning или error? тут же может быть как бизнесово ошибка так и технически. По какому принципу выбираем уровень логирования на ошибку?
//		return err
//	}
//	//почему надо ставить отдельно c.Set("Location", endpoint + ID) ? каждый раз ли это пишут в хендлере отдельно? и по самой логике не очень понял для чего возвращать слово Location и эндпоинт?
//	resp := &brokerdto.FixationResponse{ //мы возвращаем отдельно dto вместо напрямую protoResp потому что в dto есть json теги, а в protoResp нет? а если добавить?
//		FixationIDNew: protoResp.GetFixationId(),
//		//Status:        fixationStatusToString(protoResp.GetStatus()),
//		FixedAt:   formatTimestamp(protoResp.GetFixedAt()),
//		ExpiresAt: formatTimestamp(protoResp.GetExpiresAt()),
//	}
//
//	return c.Status(fiber.StatusCreated).JSON(resp)
//}

//func fixationStatusToString(status brokerv1.FixationStatus) string {
//	switch status {
//	case brokerv1.FixationStatus_FIXATION_STATUS_ACTIVE:
//		return "active"
//	case brokerv1.FixationStatus_FIXATION_STATUS_CONVERTED:
//		return "converted"
//	case brokerv1.FixationStatus_FIXATION_STATUS_EXPIRED:
//		return "expired"
//	case brokerv1.FixationStatus_FIXATION_STATUS_REMOVED:
//		return "removed"
//	default:
//		return "unspecified"
//	}
//}

//func formatTimestamp(ts *timestamppb.Timestamp) string {
//	if ts == nil {
//		return ""
//	}
//
//	return ts.AsTime().Format(time.RFC3339)
//}
