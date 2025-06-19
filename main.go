/*
date:2020-08-15
author:awake1t
*/
package main

import (
	"PortBrute/brute"
	"PortBrute/common"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"os"
	"time"
)

func main() {
    ips := flag.String("f", "ip.txt", "要爆破的ip列表")
    thread := flag.Int("t", 100, "扫描线程")
    user := flag.String("u", "user.txt", "用户名列表")
    pass := flag.String("p", "pass.txt", "密码列表")
    userPass := flag.Bool("up", false, "使用user:pass字典模式")
    flag.Parse()

    startTime := time.Now()
    userDict, _ := common.ReadUserDict(*user)
    passDict, _ := common.ReadUserDict(*pass)
    ipList := common.ReadIpList(*ips)

    color.Cyan("Thread: %d", *thread)
    color.Cyan("Number of IPs: %d", len(ipList))
    color.Cyan("Number of username dict: %d", len(userDict))
    color.Cyan("Number of password dict: %d", len(passDict))

    // 根据模式生成任务通道
    var taskChan <-chan models.Service
    if *userPass {
        taskChan = brute.GenerateTaskUserPass(ipList, userDict)
    } else {
        taskChan = brute.GenerateTask(ipList, userDict, passDict)
    }

    // 启动固定数量的工作协程池
    var wg sync.WaitGroup
    for i := 0; i < *thread; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            brute.RunBrute(taskChan)
        }()
    }

    // 等待所有扫描协程完成
    wg.Wait()
    brute.WriteToFile("\n全部扫描完成\n", "res.txt")
    color.Red("Run Time: %s\n", time.Since(startTime))
}
