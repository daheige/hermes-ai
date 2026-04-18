package config

import (
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

var (
	DB    *gorm.DB
	LOGDB *gorm.DB
)

func chooseDB(envName string) (*gorm.DB, error) {
	dsn := os.Getenv(envName)

	switch {
	case strings.HasPrefix(dsn, "postgres://"):
		// Use PostgreSQL
		return openPostgreSQL(dsn)
	case dsn != "":
		// Use MySQL
		return openMySQL(dsn)
	default:
		panic("database not found in environment variables")
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

func InitDB() {
	var err error
	DB, err = chooseDB("SQL_DSN")
	if err != nil {
		log.Fatalln("failed to initialize database: " + err.Error())
		return
	}
}

func InitLogDB() {
	if os.Getenv("LOG_SQL_DSN") == "" {
		LOGDB = DB
		return
	}

	slog.Info("using secondary database for table logs")
	var err error
	LOGDB, err = chooseDB("LOG_SQL_DSN")
	if err != nil {
		log.Fatalln("failed to initialize secondary database: " + err.Error())
		return
	}

	setDBConns(LOGDB)
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

func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}

func CloseDB() error {
	if LOGDB != DB {
		err := closeDB(LOGDB)
		if err != nil {
			return err
		}
	}

	return closeDB(DB)
}
