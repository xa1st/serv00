package main

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	Host     string
	User     string
	Password string
}

// 自己实现的Interactive方法
func passwordKeyboardInteractive(password string) ssh.KeyboardInteractiveChallenge {
	return func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		answers := make([]string, len(questions))
		for i := range answers {
			answers[i] = password
		}
		return answers, nil
	}
}

// 连接ssh
func SshConnect(server Server) (*ssh.Client, error) {
	// 创建SSH客户端配置
	config := &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			// Password方式
			ssh.Password(server.Password),
			// 键盘交互式认证，国外的一些免费主机都是关闭了PASSWORD方式，需要用这种方式来验证
			ssh.KeyboardInteractive(passwordKeyboardInteractive(server.Password)),
		},
		// 注意：生产环境中不要使用InsecureIgnoreHostKey
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return ssh.Dial("tcp", server.Host, config)
}

func main() {
	// 服务器信息
	var server Server = Server{Host: "10.6.1.6:22", User: "root", Password: "0"}
	// 开启连接
	client, err := SshConnect(server)
	if err != nil {
		fmt.Println("Failed to connect server: ", err)
		return
	}
	defer client.Close()

	// 创建一个新的session
	session, err := client.NewSession()
	if err != nil {
		fmt.Println("Failed to create session: ", err)
		return
	}
	defer session.Close()

	// 执行命令
	output, err := session.CombinedOutput("ls -l")
	if err != nil {
		fmt.Println("Failed to run command: ", err)
	}
	fmt.Println(string(output))
}
