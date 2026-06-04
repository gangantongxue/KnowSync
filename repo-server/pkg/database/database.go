// Package database 提供 MySQL 数据库连接管理和连接池配置.
package database

import (
	"fmt"

	"github.com/gangantongxue/knowsync/repo-server/pkg/config"
	"github.com/gangantongxue/knowsync/repo-server/pkg/logger"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Database 数据库连接，封装 GORM 实例及连接池配置.
type Database struct {
	DB     *gorm.DB
	Cfg    *config.Config
	Logger *logger.Logger
}

// NewDatabase 创建 MySQL 数据库连接并配置连接池.
func NewDatabase(cfg *config.Config, logger *logger.Logger) (*Database, error) {
	gormLogger := slogGorm.New(
		slogGorm.WithHandler(logger.MultiHandler),
	)

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

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return &Database{DB: db, Cfg: cfg, Logger: logger}, nil
}
