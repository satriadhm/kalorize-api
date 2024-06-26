package services

import (
	"context"
	"fmt"
	"io"
	"kalorize-api/app/models"
	"kalorize-api/app/repositories"
	"kalorize-api/utils"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type adminService struct {
	gymRepo       repositories.GymRepository
	userRepo      repositories.UserRepository
	gymKode       repositories.KodeGymRepository
	gymUsedCode   repositories.UsedCodeRepository
	makananRepo   repositories.MakananRepository
	franchiseRepo repositories.FranchiseRepository
}

func NewAdminService(db *gorm.DB) AdminService {
	return &adminService{
		userRepo:      repositories.NewDBUserRepository(db),
		gymRepo:       repositories.NewDBGymRepository(db),
		gymKode:       repositories.NewDBKodeGymRepository(db),
		gymUsedCode:   repositories.NewDBUsedCodeRepository(db),
		makananRepo:   repositories.NewDBMakananRepository(db),
		franchiseRepo: repositories.NewDBFranchiseRepository(db),
	}
}

func (service *adminService) RegisterGym(token string, registGymRequest utils.GymRequest, photoRequest utils.UploadedPhoto) utils.Response {
	var response utils.Response
	adminEmail, err := utils.ParseDataEmail(token)
	if adminEmail == "" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	admin, err := service.userRepo.GetUserByEmail(adminEmail)
	if admin.Role != "admin" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}

	gym := models.Gym{
		IdGym:      uuid.New(),
		NamaGym:    registGymRequest.NamaGym,
		AlamatGym:  registGymRequest.AlamatGym,
		Latitude:   registGymRequest.Latitude,
		Longitude:  registGymRequest.Longitude,
		LinkGoogle: registGymRequest.LinkGoogle,
	}

	// Initialize Firebase app
	opt := option.WithCredentialsFile("config/credentials.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to initialize Firebase app"
		response.Data = nil
		return response
	}

	// Initialize Firebase Storage client
	client, err := app.Storage(context.Background())
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to initialize Firebase Storage client"
		response.Data = nil
		return response
	}

	// Specify the path within the bucket where the file should be stored
	storagePath := fmt.Sprintf("images/%s", photoRequest.Handler.Filename)

	// Open a new reader for the file
	reader := photoRequest.File

	// Get the bucket handle from the client
	bucket, err := client.Bucket("kalorize-71324.appspot.com")
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to get bucket handle from the client"
		response.Data = nil
		return response
	}

	// Initialize the writer for the file
	wc := bucket.Object(storagePath).NewWriter(context.Background())
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}

	// Upload the file to Firebase Storage
	if _, err := io.Copy(wc, reader); err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to upload file to Firebase Storage"
		response.Data = nil
		return response
	}

	// Close the writer after copying
	if err := wc.Close(); err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to close Firebase Storage writer"
		response.Data = nil
		return response
	}

	// Set gym properties
	gym.PhotoGym = photoRequest.Alias + filepath.Ext(photoRequest.Handler.Filename)
	gym.PhotoUrl = fmt.Sprintf("https://storage.googleapis.com/kalorize-71324.appspot.com/%s", storagePath)
	err = service.gymRepo.CreateNewGym(gym)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to create gym"
		response.Data = nil
		return response
	}
	response.StatusCode = 200
	response.Messages = "Success"
	response.Data = gym
	return response
}

func (service *adminService) RegisterFranchise(bearerToken string, registerFranchiseRequest utils.FranchiseRequest) utils.Response {
	var response utils.Response
	adminEmail, err := utils.ParseDataEmail(bearerToken)
	if adminEmail == "" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	admin, err := service.userRepo.GetUserByEmail(adminEmail)
	if admin.Role != "admin" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerFranchiseRequest.PasswordFranchise), bcrypt.DefaultCost)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Password hashing failed"
		response.Data = nil
		return response
	}

	franchise := models.Franchise{
		IdFranchise:        uuid.New(),
		NamaFranchise:      registerFranchiseRequest.NamaFranchise,
		EmailFranchise:     registerFranchiseRequest.EmailFranchise,
		LongitudeFranchise: registerFranchiseRequest.LongitudeFranchise,
		LatitudeFranchise:  registerFranchiseRequest.LatitudeFranchise,
		LokasiFranchise:    registerFranchiseRequest.LokasiFranchise,
		FotoFranchise:      registerFranchiseRequest.FotoFranchise,
		PasswordFranchise:  string(hashedPassword),
		NoTeleponFranchise: registerFranchiseRequest.NoTeleponFranchise,
	}
	err = service.franchiseRepo.CreateFranchise(franchise)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to create franchise"
		response.Data = nil
		return response
	}
	response.StatusCode = 200
	response.Messages = "Success"
	response.Data = franchise
	return response
}

func (service *adminService) RegisterMakanan(bearerToken string, registMakananRequest utils.MakananRequest) utils.Response {
	var response utils.Response
	adminEmail, err := utils.ParseDataEmail(bearerToken)
	if adminEmail == "" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	admin, err := service.userRepo.GetUserByEmail(adminEmail)
	if admin.Role != "admin" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}

	id := utils.GenerateIdMakanan(registMakananRequest.Nama)
	makanan := models.Makanan{
		IdMakanan:     id,
		Nama:          registMakananRequest.Nama,
		Kalori:        registMakananRequest.Kalori,
		Protein:       registMakananRequest.Protein,
		ListFranchise: strings.Join(registMakananRequest.ListFranchise, ", "),
		Bahan:         strings.Join(registMakananRequest.Bahan, ", "),
		CookingStep:   strings.Join(registMakananRequest.CookingStep, "., "),
	}
	err = service.makananRepo.CreateMakanan(makanan)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to create makanan"
		response.Data = nil
		return response
	}
	response.StatusCode = 200
	response.Messages = "Success"
	response.Data = makanan
	return response
}

func (service *adminService) GenerateGymToken(bearerToken string, idGym uuid.UUID) utils.Response {
	var response utils.Response
	adminEmail, err := utils.ParseDataEmail(bearerToken)
	if adminEmail == "" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	admin, err := service.userRepo.GetUserByEmail(adminEmail)
	if admin.Role != "admin" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}

	gym, err := service.gymRepo.GetGymById(idGym)
	if err != nil {
		response.StatusCode = 404
		response.Messages = "Gym not found"
		response.Data = nil
		return response
	}

	kodeGym := models.KodeGym{
		IdKodeGym:   uuid.New(),
		KodeGym:     utils.GenerateKodeGym(gym.NamaGym),
		IdGym:       gym.IdGym,
		ExpiredTime: time.Now().AddDate(0, 0, 7),
	}

	response.StatusCode = 200
	response.Messages = "Success"
	response.Data = kodeGym
	return response
}

func (service *adminService) RegisterUser(bearerToken string, registerUserRequest utils.UserRequest, photoRequest utils.UploadedPhoto) utils.Response {
	var response utils.Response
	adminEmail, err := utils.ParseDataEmail(bearerToken)
	if adminEmail == "" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	admin, err := service.userRepo.GetUserByEmail(adminEmail)
	if admin.Role != "admin" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUserRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Password hashing failed"
		response.Data = nil
		return response
	}

	user := models.User{
		IdUser:       uuid.New(),
		Email:        registerUserRequest.Email,
		Fullname:     registerUserRequest.Fullname,
		Umur:         registerUserRequest.Umur,
		BeratBadan:   registerUserRequest.BeratBadan,
		TinggiBadan:  registerUserRequest.TinggiBadan,
		JenisKelamin: registerUserRequest.JenisKelamin,
		FrekuensiGym: registerUserRequest.FrekuensiGym,
		TargetKalori: registerUserRequest.TargetKalori,
		NoTelepon:    registerUserRequest.NoTelepon,
		Password:     string(hashedPassword),
		Role:         registerUserRequest.Role,
	}

	opt := option.WithCredentialsFile("config/credentials.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to initialize Firebase app"
		response.Data = nil
		return response
	}

	client, err := app.Storage(context.Background())
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to initialize Firebase Storage client"
		response.Data = nil
		return response
	}

	storagePath := fmt.Sprintf("images/%s", registerUserRequest.Foto)
	reader := strings.NewReader(registerUserRequest.Foto)
	bucket, err := client.Bucket("kalorize-71324.appspot.com")
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to get bucket handle from the client"
		response.Data = nil
		return response
	}

	wc := bucket.Object(storagePath).NewWriter(context.Background())
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	if _, err := io.Copy(wc, reader); err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to upload file to Firebase Storage"
		response.Data = nil
		return response
	}

	if err := wc.Close(); err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to close Firebase Storage writer"
		response.Data = nil
		return response
	}

	user.Foto = registerUserRequest.Foto
	user.FotoUrl = fmt.Sprintf("https://storage.googleapis.com/kalorize-71324.appspot.com/%s", storagePath)

	err = service.userRepo.CreateNewUser(user)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to create user"
		response.Data = nil
		return response
	}
	response.StatusCode = 200
	response.Messages = "Success"
	response.Data = user
	return response
}

func (service *adminService) GetAllUser(bearerToken string) utils.Response {
	var response utils.Response
	adminEmail, err := utils.ParseDataEmail(bearerToken)
	if adminEmail == "" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	admin, err := service.userRepo.GetUserByEmail(adminEmail)
	if admin.Role != "admin" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}

	users, err := service.userRepo.GetAllUser()
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to get all user"
		response.Data = nil
		return response
	}
	response.StatusCode = 200
	response.Messages = "Success"
	response.Data = users
	return response
}

func (service *adminService) GetUserById(bearerToken string, id uuid.UUID) utils.Response {
	var response utils.Response
	adminEmail, err := utils.ParseDataEmail(bearerToken)
	if adminEmail == "" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	admin, err := service.userRepo.GetUserByEmail(adminEmail)
	if admin.Role != "admin" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}

	user, err := service.userRepo.GetUserById(id)
	if err != nil {
		response.StatusCode = 404
		response.Messages = "User not found"
		response.Data = nil
		return response
	}
	response.StatusCode = 200
	response.Messages = "Success"
	response.Data = user
	return response
}

func (service *adminService) UpdateUser(bearerToken string, id uuid.UUID, updateUserRequest utils.UserRequest) utils.Response {
	var response utils.Response
	adminEmail, err := utils.ParseDataEmail(bearerToken)
	if adminEmail == "" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	admin, err := service.userRepo.GetUserByEmail(adminEmail)
	if admin.Role != "admin" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}

	user, err := service.userRepo.GetUserById(id)
	if err != nil {
		response.StatusCode = 404
		response.Messages = "User not found"
		response.Data = nil
		return response
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateUserRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Password hashing failed"
		response.Data = nil
		return response
	}

	if updateUserRequest.Email != "" {
		user.Email = updateUserRequest.Email
	}

	if updateUserRequest.Fullname != "" {
		user.Fullname = updateUserRequest.Fullname
	}

	if updateUserRequest.Umur != 0 {
		user.Umur = updateUserRequest.Umur
	}

	if updateUserRequest.BeratBadan != 0 {
		user.BeratBadan = updateUserRequest.BeratBadan
	}

	if updateUserRequest.TinggiBadan != 0 {
		user.TinggiBadan = updateUserRequest.TinggiBadan
	}

	if updateUserRequest.FrekuensiGym < 4 && updateUserRequest.FrekuensiGym > -1 {
		user.FrekuensiGym = updateUserRequest.FrekuensiGym
	}

	if updateUserRequest.TargetKalori < 4 && updateUserRequest.TargetKalori > -1 {
		user.TargetKalori = updateUserRequest.TargetKalori
	}

	if updateUserRequest.NoTelepon != "" {
		user.NoTelepon = updateUserRequest.NoTelepon
	}

	if updateUserRequest.Role != "" {
		user.Role = updateUserRequest.Role
	}

	if updateUserRequest.Password != "" {
		user.Password = string(hashedPassword)
	}

	err = service.userRepo.UpdateUser(user)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to update user"
		response.Data = nil
		return response
	}
	response.StatusCode = 200
	response.Messages = "Success"
	response.Data = user
	return response
}

func (service *adminService) DeleteUser(bearerToken string, id uuid.UUID) utils.Response {
	var response utils.Response
	adminEmail, err := utils.ParseDataEmail(bearerToken)
	if adminEmail == "" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}
	admin, err := service.userRepo.GetUserByEmail(adminEmail)
	if admin.Role != "admin" || err != nil {
		response.StatusCode = 401
		response.Messages = "Unauthorized"
		response.Data = nil
		return response
	}

	err = service.userRepo.DeleteUser(id)
	if err != nil {
		response.StatusCode = 500
		response.Messages = "Failed to delete user"
		response.Data = nil
		return response
	}
	response.StatusCode = 200
	response.Messages = "Success"
	return response
}

type AdminService interface {
	RegisterGym(bearerToken string, registGymRequest utils.GymRequest, photoRequest utils.UploadedPhoto) utils.Response
	RegisterFranchise(bearerToken string, registFranchiseRequest utils.FranchiseRequest) utils.Response
	RegisterMakanan(bearerToken string, registMakananRequest utils.MakananRequest) utils.Response
	RegisterUser(bearerToken string, registerUserRequest utils.UserRequest, photoRequest utils.UploadedPhoto) utils.Response
	GenerateGymToken(bearerToken string, idGym uuid.UUID) utils.Response
	GetAllUser(bearerToken string) utils.Response
	GetUserById(bearerToken string, id uuid.UUID) utils.Response
	UpdateUser(bearerToken string, id uuid.UUID, updateUserRequest utils.UserRequest) utils.Response
	DeleteUser(bearerToken string, id uuid.UUID) utils.Response
}
