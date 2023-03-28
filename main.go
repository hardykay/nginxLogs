package main

import (
	"errors"
	"fmt"
	"sync"
)

// 忽略的请求类型
var ignores = []string{"OPTIONS"}

// IsIgnore 忽略的结果会返回下面这个错误
var IsIgnore = errors.New("ignore")

func main() {
	var (
		chSize     = 20000 // 解析通道大小
		insertSize = 4369  // 批量插入允许插入的最大行，MySql最大允许65535个占位符，也就是4369条
		handelNum  = 20    // 处理日志的协程数量
		insertNum  = 5     // 批量插入协程数量
	)
	// 先走迁移创建表
	DBInit()
	if conf.Reset && DBObj.Migrator().HasTable(&RequestInfo{}) {
		err := DBObj.Migrator().DropTable(&RequestInfo{})
		if err != nil {
			panic(err)
		}
	}
	if err := DBObj.Migrator().CreateTable(&RequestInfo{}); err != nil {
		panic(fmt.Errorf("迁移表失败: %w", err))
	}
	var ch = make(chan string, chSize)
	var data = make(chan RequestInfo, chSize)
	var wg sync.WaitGroup
	// 读取日志程序
	wg.Add(1)
	go func() {
		defer wg.Done()
		Read(ch)
	}()
	// 分析日志程序
	var hold = make(chan struct{}, handelNum)
	wg.Add(handelNum)
	for i := 0; i < handelNum; i++ {
		go func() {
			defer func() {
				hold <- struct{}{}
			}()
			defer wg.Done()
			HandelData(ch, data)
		}()
	}
	// 批量新增
	for i := 0; i < insertNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Add(data, insertSize)
		}()
	}
	// 释放结束任务
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(hold)
		defer close(data)
		var num = 0
		for _ = range hold {
			num++
			if num == handelNum {
				return
			}
		}
	}()
	wg.Wait()
	fmt.Println("处理成功")
	//CountUrl()
}
