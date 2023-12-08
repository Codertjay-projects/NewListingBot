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
	incomingRoutes.Get("api/v1/orders", controllers.OrderListController)
	incomingRoutes.Post("api/v1/orders", controllers.OrderCreateController)
	incomingRoutes.Get("api/v1/symbols", controllers.GetMarketDataController)
}
