package repository

import (
	"fmt"
	"log/slog"

	"github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Repository 数据访问层.
type Repository struct {
	DB *gorm.DB
}

// NewRepository 创建数据库连接和 Repository.
func NewRepository(cfg *model.DatabaseCfg) (*Repository, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取 sql.DB 失败: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	slog.Info("数据库连接成功")

	// 自动迁移表结构
	if err := db.AutoMigrate(&ChatSession{}, &ChatMessage{}); err != nil {
		return nil, fmt.Errorf("自动迁移失败: %w", err)
	}

	return &Repository{DB: db}, nil
}
