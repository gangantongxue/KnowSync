package database

import (
	"os"
	"testing"

	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/config/model"
	"github.com/gangantongxue/knowsync/user-server/pkg/logger"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGormLoggerConfig(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		Database: model.DatabaseCfg{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "password",
			DBName:   "test_db",
		},
		Logger: model.LoggerCfg{
			Dir:        tempDir,
			Level:      "debug",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
			LocalTime:  true,
		},
		Project: model.ProjectCfg{
			Name:    "test-project",
			Version: "1.0.0",
		},
	}

	log, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("logger.NewLogger failed: %v", err)
	}
	if log == nil {
		t.Fatal("logger.NewLogger returned nil")
	}
	if log.MultiHandler == nil {
		t.Fatal("logger.MultiHandler is nil")
	}

	gormLogger := slogGorm.New(
		slogGorm.WithHandler(log.MultiHandler),
	)

	if gormLogger == nil {
		t.Fatal("slogGorm.New returned nil")
	}

	dsn := "invalid_dsn_for_testing_only"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err == nil {
		t.Log("Note: Database connection succeeded with invalid DSN (unexpected)")
	} else {
		t.Logf("Expected database connection error: %v", err)
	}

	if db != nil {
		sqlDB, err := db.DB()
		if err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}

	if db != nil && db.Config != nil && db.Config.Logger == nil {
		t.Error("gorm.Config.Logger is nil - logger was not applied correctly")
	}
}

func TestNewDatabase_Integration(t *testing.T) {
	t.Skip("Skipping integration test - requires real MySQL database connection")

	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		Database: model.DatabaseCfg{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "password",
			DBName:   "test_db",
		},
		Logger: model.LoggerCfg{
			Dir:        tempDir,
			Level:      "debug",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
			LocalTime:  true,
		},
		Project: model.ProjectCfg{
			Name:    "test-project",
			Version: "1.0.0",
		},
	}

	log, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("logger.NewLogger failed: %v", err)
	}

	db, err := NewDatabase(cfg, log)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	if db == nil {
		t.Fatal("NewDatabase returned nil")
	}
	if db.DB == nil {
		t.Error("db.DB is nil")
	}
	if db.Cfg == nil {
		t.Error("db.Cfg is nil")
	}
	if db.Logger == nil {
		t.Error("db.Logger is nil")
	}

	if db.Cfg != cfg {
		t.Error("db.Cfg does not match input cfg")
	}
	if db.Logger != log {
		t.Error("db.Logger does not match input logger")
	}

	if db.DB != nil {
		gormConfig := db.DB.Config
		if gormConfig == nil {
			t.Error("gorm.Config is nil")
		} else if gormConfig.Logger == nil {
			t.Error("gorm.Config.Logger is nil - logger was not applied correctly")
		}
	}
}
