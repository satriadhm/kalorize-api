package routes

import (
	"kalorize-api/domain/auth/controllers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RouteMakanan(apiv1 *echo.Group, db *gorm.DB) {
	makananController := controllers.NewMakananController(db)

	apiv1.GET("/makanan", makananController.GetAllMakanan)
	apiv1.GET("/makanan/:makananId", makananController.GetMakananById)
}
