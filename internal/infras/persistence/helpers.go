package persistence

import (
	"gorm.io/gorm"
)

// dbDialect 获取数据库方言名称
func dbDialect(db *gorm.DB) string {
	return db.Dialector.Name()
}

// isPostgreSQL 是否为PostgreSQL
func isPostgreSQL(db *gorm.DB) bool {
	return dbDialect(db) == "postgres"
}

// groupCol 获取group列名(处理保留字)
func groupCol(db *gorm.DB) string {
	if isPostgreSQL(db) {
		return `"group"`
	}
	return "`group`"
}

// keyCol 获取key列名(处理保留字)
func keyCol(db *gorm.DB) string {
	if isPostgreSQL(db) {
		return `"key"`
	}
	
	return "`key`"
}

// trueVal 获取布尔值true的SQL表示
func trueVal(db *gorm.DB) string {
	if isPostgreSQL(db) {
		return "true"
	}

	return "1"
}

// randomOrder 获取随机排序SQL
func randomOrder(db *gorm.DB) string {
	if isPostgreSQL(db) {
		return "RANDOM()"
	}

	return "RAND()"
}

// ifnullFunc 获取ifnull函数名
func ifnullFunc(db *gorm.DB) string {
	if isPostgreSQL(db) {
		return "COALESCE"
	}

	return "ifnull"
}

// dateFormatSQL 获取日期格式化SQL
func dateFormatSQL(db *gorm.DB) string {
	if isPostgreSQL(db) {
		return "TO_CHAR(date_trunc('day', to_timestamp(created_at)), 'YYYY-MM-DD') as day"
	}

	return "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as day"
}
