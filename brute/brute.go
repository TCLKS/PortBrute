package brute

import (
	"PortBrute/common"
	"PortBrute/models"
	"PortBrute/plugins"
	"fmt"
	"github.com/fatih/color"
	"os"
	"strconv"
	"strings"
	"sync"
	"github.com/cheggaaa/pb/v3"
	"time"
)

var bruteResult map[string]models.Service

func saveRes(target models.Service, hash string) {
    if bruteResult == nil {
        bruteResult = make(map[string]models.Service)
    }
    if _, ok := bruteResult[hash]; !ok {
        color.Cyan("[+] %s %d %s %s\n", target.Ip, target.Port, target.UserName, target.PassWord)
        s := fmt.Sprintf("[+] %s %d %s %s\n", target.Ip, target.Port, target.UserName, target.PassWord)
        WriteToFile(s, "res.txt")
        bruteResult[hash] = target
    }
}

// RunBrute 从任务通道中读取任务并执行爆破
func RunBrute(taskChan <-chan models.Service) {
    for target := range taskChan {
        protocol := strings.ToUpper(target.Protocol)
        var key string
        if protocol == "REDIS" || protocol == "FTP" || protocol == "SNMP" ||
            protocol == "POSTGRESQL" || protocol == "SSH" {
            key = fmt.Sprintf("%v-%v-%v", target.Ip, target.Port, target.Protocol)
        } else {
            key = fmt.Sprintf("%v-%v-%v", target.Ip, target.Port, target.UserName)
        }
        h := common.MakeTaskHash(key)
        if _, ok := bruteResult[h]; ok {
            continue
        }
        err, success := plugins.ScanFuncMap[protocol](
            target.Ip, strconv.Itoa(target.Port),
            target.UserName, target.PassWord)
        if err == nil && success {
            saveRes(target, h)
        }
    }
}

// GenerateTaskUserPass 使用 user:pass 列表生成任务通道
func GenerateTaskUserPass(addr []models.IpAddr, userList []string) <-chan models.Service {
    taskChan := make(chan models.Service, 100)
    go func() {
        defer close(taskChan)
        for _, up := range userList {
            parts := strings.Split(up, ":")
            if len(parts) != 2 {
                continue
            }
            user := parts[0]; pass := parts[1]
            for _, ip := range addr {
                taskChan <- models.Service{
                    Ip:       ip.Ip,
                    Port:     ip.Port,
                    Protocol: ip.Protocol,
                    UserName: user,
                    PassWord: pass,
                }
            }
        }
    }()
    return taskChan
}

// GenerateTask 使用用户名、密码列表生成任务通道（并包含空密码尝试）
func GenerateTask(addr []models.IpAddr, userList []string, passList []string) <-chan models.Service {
    taskChan := make(chan models.Service, 100)
    go func() {
        defer close(taskChan)
        // 针对部分协议尝试空凭证
        for _, ip := range addr {
            if ip.Protocol == "REDIS" || ip.Protocol == "FTP" ||
               ip.Protocol == "POSTGRESQL" || ip.Protocol == "SSH" {
                taskChan <- models.Service{
                    Ip:       ip.Ip,
                    Port:     ip.Port,
                    Protocol: ip.Protocol,
                    UserName: "",
                    PassWord: "",
                }
            }
        }
        // 普通的用户密码组合
        for _, user := range userList {
            for _, pass := range passList {
                for _, ip := range addr {
                    taskChan <- models.Service{
                        Ip:       ip.Ip,
                        Port:     ip.Port,
                        Protocol: ip.Protocol,
                        UserName: user,
                        PassWord: pass,
                    }
                }
            }
        }
    }()
    return taskChan
}

// WriteToFile 将结果追加到文件
func WriteToFile(content, filename string) {
    fd, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    fd.Write([]byte(content))
}
