package routes

import (
	"NewListingBot/controllers"
	"github.com/gofiber/fiber/v2"
)

/*This contains all the routes on the user-services Combined
 */

func HttpRoutes(app *fiber.App) {

	Routers(app)
}

func Routers(incomingRoutes *fiber.App) {
	incomingRoutes.Get("/orders", controllers.ListOrdersController)
}
