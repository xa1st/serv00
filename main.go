package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// ServerConfig 定义服务器配置的结构
type ServerConfig struct {
	Host string `json:"host"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

// 返回当前ip地址
func getIp() (ret string, err error) {
	url := "https://api.ipify.org/"
	// 创建一个http.Client实例，并设置超时时间
	client := http.Client{
		Timeout: 10 * time.Second, // 设置超时时间为10秒
	}
	// 使用http.Client发送GET请求
	resp, err := client.Get(url)
	if err != nil {
		return ret, err
	}
	// 结束时关闭响应体
	defer resp.Body.Close()
	// 读取响应体内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	return string(body), err
}

// 登陆ssh并且执行命令
func sshcmd(server ServerConfig, results chan string, cmdip string) {
	defer func() {
		results <- "done"
	}()
	fmt.Println(server)
	client, err := ssh.Dial("tcp", server.Host, &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(server.Pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		results <- fmt.Sprintf("用户 %s 连接到 %s 失败: %s", server.User, server.Host, err)
		return
	}
	defer client.Close()
	// 拿起会话
	session, err := client.NewSession()
	if err != nil {
		results <- fmt.Sprintf("用户名：%s，创建会话失败:%s", server.User, err)
		return
	}
	defer session.Close()
	// 当前时间
	cmdtime := time.Now().Format("20060102150405")
	// 要执行的命令
	command := fmt.Sprintf("echo \"%s >> ./logs/%s.log\"", cmdip, cmdtime)
	// 直接执行命令，不需要返回值
	_, err = session.CombinedOutput(command)
	if err != nil {
		results <- fmt.Sprintf("用户名：%s，执行命令失败:%s", server.User, err)
	}
}

func main() {
	// 从环境变量中获取SERVER的值
	serverData := os.Getenv("SERVER")
	if serverData == "" {
		fmt.Println("SERVER环境变量未设置")
		return
	}
	// 解析一下JSON数据
	var servers []ServerConfig
	err := json.Unmarshal([]byte(serverData), &servers)
	if err != nil {
		fmt.Println("解析SERVER环境变量失败:", err)
		return
	}
	// 定义通道来收集命令输出
	results := make(chan string)
	// 获取当前脚本运行时服务器IP
	cmdip, _ := getIp()
	// 定义协程
	var wg sync.WaitGroup
	// 执行主体
	for _, server := range servers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 这里登陆ssh并执行对应命令
			sshcmd(server, results, cmdip)
		}()
	}
	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(results)
	}()
	// 从通道中读取命令输出
	for result := range results {
		if result != "done" {
			fmt.Println(result)
		}
	}
}
