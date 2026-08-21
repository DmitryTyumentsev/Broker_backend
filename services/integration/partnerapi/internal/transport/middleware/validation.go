package middleware

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/dto"
	"errors"
	"reflect"
	"strings"

	validate "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

const validatedBodyLocalKey = "request.validated_body"

type RequestValidator interface {
	Struct(s any) error
}

func NewRequestValidator() *validate.Validate {
	validator := validate.New()
	validator.RegisterTagNameFunc(jsonTagName)

	return validator
} //верно понял что *fiber.Ctx нужен для приема запроса только? не понял немного про cli, worker, test, это же все grpc, транспорт или как? почему *fiber.Ctx не будет у них?

func ValidateJSON[T any](validator RequestValidator) fiber.Handler { //в T мы задаем dto. А на вход передаем другой тип. Разве суть дженериков не в том что мы передаем на вход тот же тип что и в T? а то тут два разных типа выходит. Не понимаю зачем нам тогда дженерики вообще, они же задуманы как проверка типа если верно понял?
	return func(c *fiber.Ctx) error {
		var req T

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Code:    fiber.StatusUnprocessableEntity,
				Message: "invalid request body",
			})
		}

		if validator != nil { //не понимаю зачем нам валидатор когда валидацию(например обработка required, uuid в dto) лежит на c.BodyParser. как и парсинг req в dto, то есть в моей картине вообще все делает fiber
			if err := validator.Struct(req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
					Code:    fiber.StatusBadRequest,
					Message: "validation failed",
					Fields:  validationFields(err),
				})
			}
		}

		c.Locals(validatedBodyLocalKey, req) //что значит value ...interface{} у методов? не понимаю суть c.Locals, почему мы его используем когда все на транспортном уровне можно передавать через *fiber.Ctx? верно понял что в fiber.Ctx кладут req и далее вытаскивают хэдеры/body/query параметры? и еще момент важный - что именно храним в c.Locals? это прям чисто связка миддлвар - хендлер? данные(req) у нас же хранятся в *fiber.Ctx. Мы их отдаем в миддлвар вадидации и дальше передаем их в клиент. Где тут c.Locals и зачем? и что в нем? или он вообще нужен чтобы просто задать ключ для данных(например body или хэдеров). то есть весь запрос это req, а что-то конкретное это константа через c.Locals?
		return c.Next()
	}
}

func ValidatedBody[T any](c *fiber.Ctx) (T, bool) { //по c.UserContext - timeout/deadline кладутся же в context.Context или я что-то не понимаю? зачем он нужен, где его используем?
	if c == nil {
		var zero T
		return zero, false
	}

	req, ok := c.Locals(validatedBodyLocalKey).(T) //разве c.Locals не мапа? когда задали связку и вызываем метод передавая ключ, получаем значение?
	return req, ok
}

func validationFields(err error) []dto.Field {
	var validationErrors validate.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil
	}

	fields := make([]dto.Field, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		fields = append(fields, dto.Field{
			Field:   fieldErr.Field(),
			Message: fieldErr.Tag(),
		})
	}

	return fields
}

func jsonTagName(field reflect.StructField) string {
	name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
	switch name {
	case "":
		return field.Name
	case "-":
		return ""
	default:
		return name
	}
}
