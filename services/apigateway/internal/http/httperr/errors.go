package httperr

import "github.com/gofiber/fiber/v2"

func WriteGRPCError(c *fiber.Ctx, err error) error {
	 = convertGRPCError(err)

}