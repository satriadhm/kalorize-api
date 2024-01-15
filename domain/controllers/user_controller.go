package controllers

import (
	"kalorize-api/domain/services"
	"kalorize-api/utils"
	"net/http"
	"strings"
	"time"

	vl "github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type UserController struct {
	userService services.UserService
	validate    vl.Validate
}

func NewUserController(db *gorm.DB) UserController {
	service := services.NewUserService(db)
	controller := UserController{
		userService: service,
		validate:    *vl.New(),
	}
	return controller
}

func (controller *UserController) EditUser(c echo.Context) error {
	authorizationHeader := c.Request().Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return c.JSON(401, "Unauthorized")
	}
	token := strings.TrimPrefix(authorizationHeader, "Bearer ")
	type payload struct {
		NamaUser  string `json:"namaUser"`
		EmailUser string `json:"emailUser"`
		NoTelepon string `json:"noTelepon"`
	}

	payloadValidator := new(payload)
	if err := c.Bind(payloadValidator); err != nil {
		return c.JSON(400, err.Error())
	}

	if err := controller.validate.Struct(payloadValidator); err != nil {
		return c.JSON(400, err.Error())
	}
	var editUserPayload utils.UserRequest = utils.UserRequest{
		Fullname:  payloadValidator.NamaUser,
		Email:     payloadValidator.EmailUser,
		NoTelepon: payloadValidator.NoTelepon,
	}

	response := controller.userService.EditUser(token, editUserPayload)
	return c.JSON(response.StatusCode, response)
}

func (controller *UserController) EditPassword(c echo.Context) error {
	authorizationHeader := c.Request().Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return c.JSON(401, "Unauthorized")
	}
	token := strings.TrimPrefix(authorizationHeader, "Bearer ")
	type payload struct {
		OldPassword              string `json:"oldPassword" validate:"required"`
		NewPassword              string `json:"newPassword" validate:"required"`
		PasswordConfirmationUser string `json:"passwordConfirmation" validate:"required"`
	}
	payloadValidator := new(payload)
	if err := c.Bind(payloadValidator); err != nil {
		return c.JSON(400, err.Error())
	}

	if err := controller.validate.Struct(payloadValidator); err != nil {
		return c.JSON(400, err.Error())
	}
	var editPasswordPayload utils.UserRequest = utils.UserRequest{
		Password:             payloadValidator.NewPassword,
		PasswordConfirmation: payloadValidator.PasswordConfirmationUser,
	}

	response := controller.userService.EditPassword(token, editPasswordPayload, payloadValidator.OldPassword)
	return c.JSON(response.StatusCode, response)
}

func (controller *UserController) EditPhoto(c echo.Context) error {
	authorizationHeader := c.Request().Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return c.JSON(401, "Unauthorized")
	}
	token := strings.TrimPrefix(authorizationHeader, "Bearer ")
	// ParseMultipartForm with a maximum of 1024 bytes
	if err := c.Request().ParseMultipartForm(1024); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	alias := c.Request().FormValue("alias")
	uploadedFile, handler, err := c.Request().FormFile("file")
	if err != nil {
		return c.JSON(400, err.Error())
	}

	photoRequest := utils.UploadedPhoto{
		Alias:   alias,
		File:    uploadedFile,
		Handler: handler,
	}

	response := controller.userService.EditPhoto(token, photoRequest)
	return c.JSON(response.StatusCode, response)
}

func (controller *UserController) CreateHistory(c echo.Context) error {
	authorizationHeader := c.Request().Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return c.JSON(401, "Unauthorized")
	}
	token := strings.TrimPrefix(authorizationHeader, "Bearer ")
	type payload struct {
		BreakfastId   []string  `json:"breakfastId"`
		LunchId       []string  `json:"lunchId"`
		DinnerId      []string  `json:"dinnerId"`
		TotalCalories int       `json:"totalCalories"`
		TotalProtein  int       `json:"totalProtein"`
		TanggalDibuat time.Time `json:"tanggalDibuat"`
	}

	payloadValidator := new(payload)
	if err := c.Bind(payloadValidator); err != nil {
		return c.JSON(400, err.Error())
	}

	if err := controller.validate.Struct(payloadValidator); err != nil {
		return c.JSON(400, err.Error())
	}
	var historyPayload utils.HistoryRequest = utils.HistoryRequest{
		IdBreakfast:   strings.Join(payloadValidator.BreakfastId, ","),
		IdLunch:       strings.Join(payloadValidator.LunchId, ","),
		IdDinner:      strings.Join(payloadValidator.DinnerId, ","),
		TotalKalori:   payloadValidator.TotalCalories,
		TotalProtein:  payloadValidator.TotalProtein,
		TanggalDibuat: payloadValidator.TanggalDibuat,
	}

	response := controller.userService.CreateHistory(token, historyPayload)
	return c.JSON(response.StatusCode, response)
}

func (controller *UserController) GetHistoryBaseDateTime(c echo.Context) error {
	authorizationHeader := c.Request().Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return c.JSON(401, "Unauthorized")
	}
	dateString := c.Param("TanggalDibuat")
	date, err := utils.StringToDate(dateString)
	if err != nil {
		return c.JSON(400, err.Error())
	}
	token := strings.TrimPrefix(authorizationHeader, "Bearer ")
	response := controller.userService.GetHistory(token, date)
	return c.JSON(response.StatusCode, response)
}
