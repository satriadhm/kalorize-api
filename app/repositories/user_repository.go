package repositories

import (
	"kalorize-api/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dbUser struct {
	Conn *gorm.DB
}

func (dbAuth *dbUser) GetToken() string {
	return "token"
}

func (db *dbUser) GetUserById(id uuid.UUID) (models.User, error) {
	var user models.User
	err := db.Conn.Where("id_user = ?", id).First(&user).Error
	return user, err
}

func (db *dbUser) CreateNewUser(user models.User) error {
	return db.Conn.Create(&user).Error
}

func (db *dbUser) GetUserByUsername(username string) (models.User, error) {
	var user models.User
	err := db.Conn.Where("full_name = ?", username).First(&user).Error
	return user, err
}

func (db *dbUser) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := db.Conn.Where("email = ?", email).First(&user).Error
	return user, err
}

func (db *dbUser) FindReferalCodeIfExist(code string) bool {
	var user models.User
	err := db.Conn.Where("referal_code = ?", code).First(&user).Error
	return err == nil
}

func (db *dbUser) UpdateUser(user models.User) error {
	err := db.Conn.Save(&user).Error
	if err != nil {
		return err // Mengembalikan error yang terjadi saat menyimpan data
	}
	return nil
}

func (db *dbUser) GetAllUser() ([]models.User, error) {
	var users []models.User
	err := db.Conn.Find(&users).Error
	return users, err
}

func (db *dbUser) DeleteUser(id uuid.UUID) error {
	var user models.User
	err := db.Conn.Where("id_user = ?", id).Delete(&user).Error
	return err
}

type UserRepository interface {
	GetToken() string
	GetAllUser() ([]models.User, error)
	CreateNewUser(user models.User) error
	DeleteUser(id uuid.UUID) error
	GetUserByUsername(username string) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	FindReferalCodeIfExist(code string) bool
	UpdateUser(user models.User) error
	GetUserById(id uuid.UUID) (models.User, error)
}

func NewDBUserRepository(conn *gorm.DB) *dbUser {
	return &dbUser{Conn: conn}
}
