package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBObj *gorm.DB

func DBInit() {
	var err error
	DBObj, err = ConnMysql()
	if err != nil {
		panic("打开数据库错误" + err.Error())
	}
	DBObj.Logger = DBObj.Logger.LogMode(logger.Info)
}

func ConnMysql() (*gorm.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", conf.User, conf.Password, conf.Host, conf.Port, conf.Dbname, conf.Params)

	return gorm.Open(mysql.Open(dataSourceName), &gorm.Config{
		SkipDefaultTransaction: true, //初始化时禁用它，这将获得大约 30%+ 性能提升。
		//FullSaveAssociations: true,//会关联更新子表数据
	})
}
