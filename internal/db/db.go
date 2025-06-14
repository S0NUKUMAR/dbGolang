package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBService struct {
	Db *gorm.DB
}

func NewDBService(dsn string) (*DBService, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &DBService{Db: db}, nil
}

func (s *DBService) Connect(dsn string) error {
	var err error
	s.Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}

func (s *DBService) AutoMigrate(models ...interface{}) error {
	return s.Db.AutoMigrate(models...)
}
