package main

import "time"

type ConfStruct struct {
	User      string
	Password  string
	Host      string
	Port      int
	Dbname    string
	Params    string
	StartTime string // 只取日志在某个时间范围，开始时间，不设置不限制,格式：2006-01-02 15:04:05
	EndTime   string // 只取日志在某个时间范围，结束时间，不设置不限制
	MinTime   time.Time
	MaxTime   time.Time
	Reset     bool // 是否重置数据库表
}

var conf = ConfStruct{
	User:     "root",
	Password: "root",
	Host:     "127.0.0.1",
	Port:     3306,
	Dbname:   "test",
	Params:   "charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&loc=Local",
	//StartTime: "2023-02-27 11:45:57",
	//EndTime:   "2023-04-27 11:45:57",
	Reset: true,
}

func init() {
	var err error
	loc, _ := time.LoadLocation("Local")
	if conf.StartTime != "" {
		conf.MinTime, err = time.ParseInLocation("2006-01-02 15:04:05", conf.StartTime, loc)
		if err != nil {
			panic(err)
		}
	}
	if conf.EndTime != "" {
		conf.MaxTime, err = time.ParseInLocation("2006-01-02 15:04:05", conf.EndTime, loc)
		if err != nil {
			panic(err)
		}
	}
}
