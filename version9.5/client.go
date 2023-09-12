// client.go
// 客户端
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //当前client模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       99,
	}

	//链接server

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIp, client.ServerPort))
	if err != nil {

		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn
	//返回对象
	return client
}

// 处理server回应的消息，直接显示到标准输出即可
func (this *Client) DealResponse() {

	//一旦this.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, this.conn)
}

func (this *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.群聊模式")
	fmt.Println("4.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 4 {
		this.flag = flag
		return true
	} else {
		fmt.Println(">>>>>请输入合法范围数字<<<<<")
		return false
	}

}

func (this *Client) PublicChat() {
	// 提示用户输入消息
	var chatMsg string
	fmt.Println(">>>>>请输入聊天内容,exit退出\n")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发送给服务器

		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write error:", err)
				break
			}

		}
		chatMsg = ""
		fmt.Println(">>>>>请输入聊天内容,exit退出\n")
		fmt.Scanln(&chatMsg)
	}

}

func (this *Client) UpdateName() bool {
	fmt.Println(">>>>>请输入用户名：")
	fmt.Scanln(&this.Name)

	sendMsg := "rename|" + this.Name + "\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		println("conn.Write err:")
		return false
	}
	return true
}

func (this *Client) Run() {
	for this.flag != 0 {
		for this.menu() != true {
		}
		//根据不同的模式处理不同的业务
		switch this.flag {
		case 1:
			//公聊模式
			fmt.Println("公聊模式...")
			this.PublicChat()
			//break
		case 2:
			//私聊模式
			fmt.Println("私聊模式...")

			//break

		case 3:
			//群聊模式
			fmt.Println("群聊模式...")

			//break

		case 4:
			//更新用户名
			fmt.Println("更新用户名...")
			this.UpdateName()
			//break

		}
	}
}

var serverIp string
var serverPort int

//./client -ip 192.168.56.105 -port 8888

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器Port(默认8888)")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> 连接服务器失败...")
		return
	}

	//单独开启一个goroutine去处理server的回执消息
	go client.DealResponse()

	fmt.Println(">>>>> 连接服务器成功...")

	//启动客户端的业务
	client.Run()
}
