package fixationhandlers

import (
	fixationv1 "Broker_backend/gen/fixation/v1"
	"Broker_backend/services/integration/partnerapi/internal/transport/dto/fixationdto"
	"Broker_backend/services/integration/partnerapi/internal/transport/grpc/grpcerr"
	"Broker_backend/services/integration/partnerapi/internal/transport/http/httperr"
	"Broker_backend/services/integration/partnerapi/internal/transport/middleware"
	"context"
	"errors"

	validate "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FixationHandler struct {
	logger    *zap.Logger
	fixation  FixationClient
	validator *validate.Validate
}

// Порядок аргументов одинаковый во всех хендлерах: logger, клиент, валидатор.
// Одинаковый порядок = меньше шансов перепутать местами два указателя,
// которые компилятор не различит, если типы совпадут.
func NewFixationHandler(logger *zap.Logger, fixation FixationClient, validator *validate.Validate) *FixationHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &FixationHandler{
		logger:    logger,
		fixation:  fixation,
		validator: validator,
	}
}

func (h *FixationHandler) Validate() error {
	switch {
	case h == nil:
		return errors.New("fixation handler is nil")
	case h.fixation == nil:
		return errors.New("fixation client is required")
	default:
		return nil
	}
}

type FixationClient interface {
	NewFixation(ctx context.Context, req *fixationv1.NewFixationRequest, meta *fixationdto.Meta) (
		*fixationv1.NewFixationResponse, error)
}

func (h *FixationHandler) NewFixation(c *fiber.Ctx) error {
	dtoReq, ok := middleware.ValidatedBody[fixationdto.FixationRequest](c)
	if !ok {
		return httperr.WriteBadRequest(c, "invalid request")
	}
	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return httperr.WriteBadRequest(c, "invalid request")
	}

	ctx := c.UserContext() //зачем перекладываем из fiber.Ctx в context.Context? разве нельзя fiber.Ctx дальше передавать?
	protoReq := &fixationv1.NewFixationRequest{
		FixFor:    dtoReq.FixFor.String(),
		Phone:     dtoReq.Phone,
		ProjectId: dtoReq.ProjectID.String(),
	}
	meta := &fixationdto.Meta{
		AgencyID: principal.AgencyID,
		FixBy:    principal.UserID,
	}
	protoResp, err := h.fixation.NewFixation(ctx, protoReq, meta)
	if err != nil {
		return grpcerr.WriteGRPCToHTTPError(c, err)
	}

	fixationID, err := uuid.Parse(protoResp.FixationId)
	if err != nil {
		return grpcerr.WriteGRPCToHTTPError(c, err)
	}

	dtoResp := &fixationdto.FixationResponse{
		FixationID: fixationID,
		FixedAt:    protoResp.FixedAt.AsTime(),
		ExpiresAt:  protoResp.ExpiresAt.AsTime(),
	}

	return c.Status(fiber.StatusCreated).JSON(dtoResp)
}

//
//func (h *FixationHandler) Validate() error {
//	switch {
//	case h == nil:
//		return errors.New("fixationservice handler is nil")
//	case h.grpcClient == nil:
//		return errors.New("fixationservice grpcClient is required")
//	default:
//		return h.grpcClient.Validate()
//	}
//}
//
//func (h *FixationHandler) NewFixation(c *fiber.Ctx) error { //Верно понял что согласно моему миддлвару аксесс лог, каждый(вообще каждый) вызов всех методов из цепочки очень подробно записывается и трейсится ещё плюсом?
//	bodyDTO, ok := middleware.ValidatedBody[fixationdto.FixationRequest](c)
//	if ok == false {
//		h.logger.Error("middleware.ValidatedBody error: type dto didn't match with c.Locals(validatedBodyKey)")
//		return c.JSON(grpcerr.WriteBadRequest(c, "invalid request"))
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
//	resp := &fixationdto.FixationResponse{ //мы возвращаем отдельно dto вместо напрямую protoResp потому что в dto есть json теги, а в protoResp нет? а если добавить?
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
