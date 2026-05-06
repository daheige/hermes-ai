package config

import (
	"errors"
	"log"
	"log/slog"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// InitDatabase 初始化db
func InitDatabase(sqlDSN, logSQLDSN string, debugSQL bool, maxIdleConns, maxOpenConns, maxLifetime int) (*gorm.DB, *gorm.DB) {
	// Initialize SQL Database
	db, err := initDB(sqlDSN, debugSQL, maxIdleConns, maxOpenConns, maxLifetime)
	if err != nil {
		log.Fatalln("failed to connect to database error:", err)
	}

	var logDB *gorm.DB
	if logSQLDSN == "" {
		logDB = db
	} else {
		logDB, err = initDB(logSQLDSN, debugSQL, maxIdleConns, maxOpenConns, maxLifetime)
		if err != nil {
			log.Fatalln("failed to connect to log database error:", err)
		}
	}

	return db, logDB
}

// initDB 初始化db
func initDB(dsn string, debugSQL bool, maxIdleConns, maxOpenConns, maxLifetime int) (*gorm.DB, error) {
	switch {
	case strings.HasPrefix(dsn, "postgres://"):
		// Use PostgreSQL
		db, err := openPostgreSQL(dsn, debugSQL)
		if err != nil {
			return nil, err
		}

		setDBConns(db, maxIdleConns, maxOpenConns, maxLifetime)
		return db, nil
	case dsn != "":
		// Use MySQL
		db, err := openMySQL(dsn, debugSQL)
		if err != nil {
			return nil, err
		}
		setDBConns(db, maxIdleConns, maxOpenConns, maxLifetime)
		return db, nil
	default:
		return nil, errors.New("database not found in environment variables")
	}
}

func openPostgreSQL(dsn string, debugSQL bool) (*gorm.DB, error) {
	slog.Info("using PostgreSQL as database")
	gormConfig := &gorm.Config{
		PrepareStmt: true, // precompile SQL
	}

	// fix gorm db logger
	if debugSQL {
		gormConfig.Logger = glogger.Default.LogMode(glogger.Info)
	}

	return gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), gormConfig)
}

func openMySQL(dsn string, debugSQL bool) (*gorm.DB, error) {
	slog.Info("using MySQL as database")

	gormConfig := &gorm.Config{
		PrepareStmt: true, // precompile SQL
	}

	// fix gorm db logger
	if debugSQL {
		gormConfig.Logger = glogger.Default.LogMode(glogger.Info)
	}

	return gorm.Open(mysql.Open(dsn), gormConfig)
}

func setDBConns(db *gorm.DB, maxIdleConns, maxOpenConns, maxLifetime int) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalln("failed to connect database: " + err.Error())
		return
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(maxLifetime))
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
