package database

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"goWeb/pkg/logger"
	"time"
)

func initDB()  {
	var err error
	
	//设置数据库连接信息
	config := mysql.Config{
		User:                     "root",
		Passwd:                   "root",
		Net:                      "tcp",
		Addr:                     "127.0.0.1:3306",
		DBName:                   "goblog",
		AllowNativePasswords:     true,
	}
	//准备数据库连接池
	db,err := sql.Open("msql",config.FormatDSN())
	logger.LogError(err)

	// 设置最大连接数
	db.SetMaxOpenConns(100)
	// 设置最大空闲连接数
	db.SetMaxIdleConns(25)
	// 设置每个链接的过期时间
	db.SetConnMaxLifetime(5 * time.Minute)

	//尝试连接，失败会报错
	err = db.Ping()
	logger.LogError(err)



}