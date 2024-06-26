package services

import (
	"context"
	"fmt"
	"io"
	"kalorize-api/app/models"
	"kalorize-api/app/repositories"
	"kalorize-api/formatter"
	"kalorize-api/utils"
	"path/filepath"
	"reflect"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type UserService interface {
	GetHistory(token string, date time.Time) utils.Response
	CreateHistory(token string, historyPayload utils.HistoryRequest) utils.Response
	EditUser(token string, payload utils.UserRequest) utils.Response
	EditPassword(token string, payload utils.UserRequest, oldPassword string) utils.Response
	EditPhoto(token string, payload utils.UploadedPhoto) utils.Response
}

type userService struct {
	userRepository     repositories.UserRepository
	historyRepository  repositories.HistoryRepository
	makananrRepository repositories.MakananRepository
}

func NewUserService(db *gorm.DB) UserService {
	return &userService{
		userRepository:     repositories.NewDBUserRepository(db),
		historyRepository:  repositories.NewDBHistoryRepository(db),
		makananrRepository: repositories.NewDBMakananRepository(db),
	}
}

func (service *userService) CreateHistory(token string, historyPayload utils.HistoryRequest) utils.Response {
	emailUser, err := utils.ParseDataEmail(token)
	if err != nil || emailUser == "" {
		return utils.Response{
			StatusCode: 401,
			Messages:   "Unauthorized",
			Data:       nil,
		}
	}
	user, err := service.userRepository.GetUserByEmail(emailUser)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get user",
			Data:       nil,
		}
	}
	t := time.Now()
	tanggal := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// Check if history already exists for the user on the same date
	existingHistory, err := service.historyRepository.GetHistoryByIdUserAndDate(user.IdUser, tanggal)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to check existing history",
			Data:       nil,
		}
	}

	// If history exists, update it; otherwise, create a new one
	if existingHistory != (models.History{}) {
		existingHistory.IdBreakfast = historyPayload.IdBreakfast
		existingHistory.IdLunch = historyPayload.IdLunch
		existingHistory.IdDinner = historyPayload.IdDinner
		existingHistory.TotalProtein = historyPayload.TotalProtein
		existingHistory.TotalKalori = historyPayload.TotalKalori
		err = service.historyRepository.UpdateHistory(existingHistory)
		if err != nil {
			return utils.Response{
				StatusCode: 500,
				Messages:   "Failed to update history",
				Data:       nil,
			}
		}
		return utils.Response{
			StatusCode: 200,
			Messages:   "History updated successfully",
			Data:       existingHistory,
		}
	}

	// Create new history if no existing history found for the user on the same date
	newHistory := models.History{
		IdHistory:     uuid.New(),
		IdUser:        user.IdUser,
		IdBreakfast:   historyPayload.IdBreakfast,
		IdLunch:       historyPayload.IdLunch,
		IdDinner:      historyPayload.IdDinner,
		TotalProtein:  historyPayload.TotalProtein,
		TotalKalori:   historyPayload.TotalKalori,
		TanggalDibuat: tanggal,
	}
	err = service.historyRepository.CreateHistory(newHistory)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to create history",
			Data:       nil,
		}
	}
	return utils.Response{
		StatusCode: 200,
		Messages:   "New history created successfully",
		Data:       newHistory,
	}
}

func (service *userService) GetHistory(token string, date time.Time) utils.Response {
	emailUser, err := utils.ParseDataEmail(token)
	if err != nil || emailUser == "" {
		return utils.Response{
			StatusCode: 401,
			Messages:   "Unauthorized",
			Data:       nil,
		}
	}
	user, err := service.userRepository.GetUserByEmail(emailUser)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get user",
			Data:       nil,
		}
	}

	history, err := service.historyRepository.GetHistoryByIdUserAndDate(user.IdUser, date)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get history",
			Data:       nil,
		}
	}
	breakfast, err := service.makananrRepository.GetMakananById(history.IdBreakfast)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get breakfast",
			Data:       nil,
		}
	}
	formattedBreakfast := formatter.FormatterMakananIndo(breakfast)
	lunch, err := service.makananrRepository.GetMakananById(history.IdLunch)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get lunch",
			Data:       nil,
		}
	}
	formattedLunch := formatter.FormatterMakananIndo(lunch)
	dinner, err := service.makananrRepository.GetMakananById(history.IdDinner)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get dinner",
			Data:       nil,
		}
	}
	formattedDinner := formatter.FormatterMakananIndo(dinner)
	var response utils.Response
	response.StatusCode = 200
	response.Messages = "Success"
	response.Data = map[string]interface{}{
		"breakfast":    formattedBreakfast,
		"lunch":        formattedLunch,
		"dinner":       formattedDinner,
		"totalKalori":  history.TotalKalori,
		"totalProtein": history.TotalProtein,
	}

	return response
}

func (service *userService) EditUser(token string, payload utils.UserRequest) utils.Response {
	nameUser, err := utils.ParseDataId(token)
	if err != nil && nameUser == uuid.Nil {
		return utils.Response{
			StatusCode: 401,
			Messages:   "Unauthorized",
			Data:       nil,
		}
	}
	user, err := service.userRepository.GetUserById(nameUser)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get user",
			Data:       nil,
		}
	}
	validateAndAssign(&user.Fullname, payload.Fullname)
	validateAndAssign(&user.Email, payload.Email)
	validateAndAssign(&user.NoTelepon, payload.NoTelepon)

	err = service.userRepository.UpdateUser(user)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to edit user",
			Data:       nil,
		}
	}
	return utils.Response{
		StatusCode: 200,
		Messages:   "Success",
		Data:       user,
	}
}

func validateAndAssign(target interface{}, source interface{}) {
	if source != nil {
		targetValue := reflect.ValueOf(target)
		sourceValue := reflect.ValueOf(source)

		if targetValue.Kind() == reflect.Ptr && !targetValue.IsNil() {
			if sourceValue.Kind() == reflect.Ptr {
				if sourceValue.Elem().IsValid() {
					targetValue.Elem().Set(sourceValue.Elem())
				}
			} else {
				// If source is not a pointer, and not an empty string, directly set the value
				if sourceValue.Kind() != reflect.String || sourceValue.String() != "" {
					targetValue.Elem().Set(sourceValue)
				}
			}
		}
	}
}

func (service *userService) EditPassword(token string, payload utils.UserRequest, oldPassword string) utils.Response {
	emailUser, err := utils.ParseDataEmail(token)
	if err != nil || emailUser == "" {
		return utils.Response{
			StatusCode: 401,
			Messages:   "Unauthorized",
			Data:       nil,
		}
	}
	user, err := service.userRepository.GetUserByEmail(emailUser)
	if err != nil && user.Email != emailUser {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get user",
			Data:       nil,
		}
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Old password is wrong",
			Data:       nil,
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Password hashing failed",
			Data:       nil,
		}
	}
	user.Password = string(hashedPassword)
	err = service.userRepository.UpdateUser(user)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to edit user",
			Data:       nil,
		}
	}

	return utils.Response{
		StatusCode: 200,
		Messages:   "Success",
		Data:       user,
	}
}

func (service *userService) EditPhoto(token string, payload utils.UploadedPhoto) utils.Response {
	emailUser, err := utils.ParseDataEmail(token)
	if err != nil || emailUser == "" {
		return utils.Response{
			StatusCode: 401,
			Messages:   "Unauthorized",
			Data:       nil,
		}
	}
	user, err := service.userRepository.GetUserByEmail(emailUser)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get user",
			Data:       nil,
		}
	}
	filename := payload.Handler.Filename
	if payload.Alias != "" {
		filename = fmt.Sprintf("%s%s", payload.Alias, filepath.Ext(payload.Handler.Filename))
	}

	// Initialize Firebase app
	opt := option.WithCredentialsFile("config/credentials.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to initialize Firebase app",
			Data:       nil,
		}
	}

	// Initialize Firebase Storage client
	// Initialize Firebase Storage client
	client, err := app.Storage(context.Background())
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to initialize Firebase Storage client",
			Data:       nil,
		}
	}

	// Specify the path within the bucket where the file should be stored
	storagePath := fmt.Sprintf("images/%s", filename)

	// Open a new reader for the file
	reader := payload.File

	// Get the bucket handle from the client
	bucket, err := client.Bucket("kalorize-71324.appspot.com")
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to get bucket handle from the client",
			Data:       nil,
		}
	}
	// Initialize the writer for the file
	wc := bucket.Object(storagePath).NewWriter(context.Background())
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}

	// Upload the file to Firebase Storage
	if _, err := io.Copy(wc, reader); err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to upload file to Firebase Storage",
			Data:       nil,
		}
	}

	// Close the writer after copying
	if err := wc.Close(); err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to close Firebase Storage writer",
			Data:       nil,
		}
	}

	// Set user properties
	user.Foto = payload.Alias + filepath.Ext(payload.Handler.Filename)
	user.FotoUrl = fmt.Sprintf("https://storage.googleapis.com/kalorize-71324.appspot.com/%s", storagePath)

	// Update user in the database
	err = service.userRepository.UpdateUser(user)
	if err != nil {
		return utils.Response{
			StatusCode: 500,
			Messages:   "Failed to edit user",
			Data:       nil,
		}
	}
	return utils.Response{
		StatusCode: 200,
		Messages:   "Success",
		Data:       user,
	}
}
