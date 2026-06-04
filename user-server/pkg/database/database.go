// Package database provides database connection and initialization.
package database

import (
	"fmt"

	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/logger"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Database 数据库连接封装.
type Database struct {
	DB     *gorm.DB
	Cfg    *config.Config
	Logger *logger.Logger
}

// NewDatabase 创建一个新的数据库连接.
func NewDatabase(cfg *config.Config, logger *logger.Logger) (*Database, error) {
	// 初始化日志器
	gormLogger := slogGorm.New(
		slogGorm.WithHandler(logger.MultiHandler),
	)

	// 初始化数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	// 设置数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return &Database{DB: db, Cfg: cfg, Logger: logger}, nil
}
