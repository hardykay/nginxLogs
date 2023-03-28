package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RequestInfo struct {
	Id                   int       `gorm:"column:id;AUTO_INCREMENT;PRIMARY_KEY;" json:"id"`
	RemoteAddr           string    `json:"remote_addr" gorm:"column:remote_addr"`                            // 请求IP
	RemoteUser           string    `json:"remote_user" gorm:"column:remote_user"`                            //
	TimeLocal            time.Time `json:"time_local" gorm:"column:time_local"`                              // 请求时间
	RequestTime          float64   `json:"request_time" gorm:"column:request_time"`                          // 请求上传的时间
	UpstreamResponseTime float64   `json:"upstream_response_time" gorm:"column:upstream_response_time"`      // 后端响应的时间
	Method               string    `gorm:"column:method" json:"method"`                                      // 请求方式
	Url                  string    `gorm:"column:url;Type:varchar(2000)" json:"url"`                         // 请求全路径
	Param                string    `gorm:"column:param;Type:varchar(2000)" json:"param"`                     // 请求路由带参数
	Path                 string    `gorm:"column:path" json:"path"`                                          // 请求路径
	HttpVersion          string    `gorm:"column:http_version" json:"http_version"`                          // http版本号
	Status               int       `gorm:"column:status" json:"status"`                                      // 请求状态
	BodyBytesSent        int       `gorm:"column:body_bytes_sent" json:"body_bytes_sent"`                    // 响应大小
	HttpReferer          string    `gorm:"column:http_referer;Type:varchar(2000)" json:"http_referer"`       // 远程地址
	HttpUserAgent        string    `gorm:"column:http_user_agent;Type:varchar(2000)" json:"http_user_agent"` // user_agent信息
	HttpXForwardedFor    string    `gorm:"column:http_x_forwarded_for" json:"http_x_forwarded_for"`          //
}

func (r RequestInfo) TableName() string {
	return "request"
}

func GetInfo(s string) (data RequestInfo, err error) {
	expr := `^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s-\s-\s\[(\d{2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2}\s\+\d{4})\s(\d+\.?\d*)\s(\d+\.?\d*)\]\s"(\w+)\s(.+)\sHTTP\/(\d\.\d)"\s(\d+)\s(\d+)\s"(.*?)"\s"(.*?)"\s"(.*?)"$`
	compileRegex := regexp.MustCompile(expr) // 正则表达式的分组，以括号()表示，每一对括号就是我们匹配到的一个文本，可以把他们提取出来。
	matchArr := compileRegex.FindStringSubmatch(s)
	if matchArr == nil || len(matchArr) < 13 {
		err = IsIgnore
		return
	}
	data.RemoteAddr = matchArr[1]
	// 时间格式
	timeLayout := "02/Jan/2006:15:04:05 -0700"
	data.TimeLocal, err = time.Parse(timeLayout, matchArr[2])
	if err != nil {
		return
	}
	if !conf.MaxTime.IsZero() && data.TimeLocal.After(conf.MaxTime) {
		err = IsIgnore
		return
	}
	if !conf.MinTime.IsZero() && data.TimeLocal.Before(conf.MinTime) {
		err = IsIgnore
		return
	}
	// 请求时间
	data.RequestTime, err = strconv.ParseFloat(matchArr[3], 64)
	if err != nil {
		return
	}
	// 处理时间
	data.UpstreamResponseTime, err = strconv.ParseFloat(matchArr[4], 64)
	if err != nil {
		return
	}
	data.Method = matchArr[5]
	for _, v := range ignores {
		if data.Method == v {
			err = IsIgnore
			return
		}
	}
	// 路由和参数
	data.Url = matchArr[6]
	url := strings.Split(data.Url, "?")
	// 处理路由中的数字
	re := regexp.MustCompile("\\d+")
	data.Path = re.ReplaceAllString(url[0], ":id")
	if len(url) >= 2 {
		data.Param = url[1]
	}
	// 匹配HTTP协议和版本号
	data.HttpVersion = matchArr[7]
	// 状态码
	data.Status, err = strconv.Atoi(matchArr[8])
	if err != nil {
		return
	}
	// 返回大小
	data.BodyBytesSent, err = strconv.Atoi(matchArr[9])
	if err != nil {
		return
	}
	// 匹配来源URL
	data.HttpReferer = matchArr[10]
	// 匹配User-Agent
	data.HttpUserAgent = matchArr[11]
	//
	data.HttpXForwardedFor = matchArr[12]
	return
}

var lock sync.Mutex

func HandelData(ch chan string, dbData chan RequestInfo) {
	for v := range ch {
		if len(v) == 0 {
			continue
		}
		data, err := GetInfo(v)
		if err == IsIgnore {
			continue
		}
		if err != nil {
			fmt.Println(v)
			panic(err)
		}
		dbData <- data
	}
}

func Add(data chan RequestInfo, max int) {
	dataList := make([]RequestInfo, 0, max)
	for v := range data {
		dataList = append(dataList, v)
		if len(dataList) == max {
			err := DBObj.Create(&dataList).Error
			if err != nil {
				panic(err)
			}
			dataList = dataList[0:0]
		}
	}
	// 最后一次不要漏
	if len(dataList) > 0 && len(dataList) < max {
		err := DBObj.Create(&dataList).Error
		if err != nil {
			panic(err)
		}
	}
}

type Count1 struct {
	Count  int    `json:"count"`
	Method string `gorm:"column:method" json:"method"` // 请求方式
	Url    string `gorm:"column:url" json:"url"`       // 请求路径
}

func CountUrl() {
	// 按请求分组查询
	// 先排查URL
	var r RequestInfo
	var count1 []Count1
	err := DBObj.Table(r.TableName()).Group("url,method").Select("count(*) AS count, method,url").Order("count desc").Find(&count1).Error
	if err != nil {
		panic(err)
	}
	fmt.Println("按请求路径统计##########################################################################")
	for _, v := range count1 {
		fmt.Printf("请求次数:%d，请求方法：%s，请求路径：%s\n", v.Count, v.Method, v.Url)
	}
}
