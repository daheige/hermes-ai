package config

import (
	"errors"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"

	"hermes-ai/internal/infras/env"
)

// InitDatabase 初始化db
func InitDatabase() (*gorm.DB, *gorm.DB) {
	// Initialize SQL Database
	db, err := initDB("SQL_DSN")
	if err != nil {
		log.Fatalln("failed to connect to database error:", err)
	}

	var logDB *gorm.DB
	if os.Getenv("LOG_SQL_DSN") == "" {
		logDB = db
	} else {
		logDB, err = initDB("SQL_DSN")
		if err != nil {
			log.Fatalln("failed to connect to log database error:", err)
		}
	}

	return db, logDB
}

// initDB 初始化db
func initDB(envName string) (*gorm.DB, error) {
	dsn := os.Getenv(envName)

	switch {
	case strings.HasPrefix(dsn, "postgres://"):
		// Use PostgreSQL
		db, err := openPostgreSQL(dsn)
		if err != nil {
			return nil, err
		}

		setDBConns(db)
		return db, nil
	case dsn != "":
		// Use MySQL
		db, err := openMySQL(dsn)
		if err != nil {
			return nil, err
		}
		setDBConns(db)
		return db, nil
	default:
		return nil, errors.New("database not found in environment variables")
	}
}

func openPostgreSQL(dsn string) (*gorm.DB, error) {
	slog.Info("using PostgreSQL as database")
	gormConfig := &gorm.Config{
		PrepareStmt: true, // precompile SQL
	}

	// fix gorm db logger
	if DebugSQLEnabled {
		gormConfig.Logger = glogger.Default.LogMode(glogger.Info)
	}

	return gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), gormConfig)
}

func openMySQL(dsn string) (*gorm.DB, error) {
	slog.Info("using MySQL as database")

	gormConfig := &gorm.Config{
		PrepareStmt: true, // precompile SQL
	}

	// fix gorm db logger
	if DebugSQLEnabled {
		gormConfig.Logger = glogger.Default.LogMode(glogger.Info)
	}

	return gorm.Open(mysql.Open(dsn), gormConfig)
}

func setDBConns(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalln("failed to connect database: " + err.Error())
		return
	}

	sqlDB.SetMaxIdleConns(env.Int("SQL_MAX_IDLE_CONNS", 100))
	sqlDB.SetMaxOpenConns(env.Int("SQL_MAX_OPEN_CONNS", 1000))
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(env.Int("SQL_MAX_LIFETIME", 60)))
}

// CloseDB 关闭db
func CloseDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}
