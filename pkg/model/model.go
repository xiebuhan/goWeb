package model

import (
	"goWeb/pkg/logger"
	"gorm.io/driver/mysql"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm"
)

var DB *gorm.DB

//connectDB 初始化模型
func ConnecDB() *gorm.DB  {
	var err error
	
	config := mysql.New(mysql.Config{
		DSN:"root:root@tcp(127.0.0.1:3306)/goblog?charset=utf8&parseTime=True&loc=Local",
	})
	DB,err = gorm.Open(config,&gorm.Config{
		Logger:gormlogger.Default.LogMode(gormlogger.Info),
	})
	logger.LogError(err)
	return  DB
}