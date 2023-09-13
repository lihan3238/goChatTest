// user.go
package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
	Groups []*Group
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 监听当前user channel的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线功能
func (this *User) Online() {
	//用户上线，将用户加入到OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

// 用户下线功能
func (this *User) Offline() {
	//用户下线，将用户从OnlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户下线消息
	this.server.BroadCast(this, "已下线")
}

// 给当前User对应的客户端发消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户都有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "在线...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]
		//判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名：" + this.Name + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式：to|张三|消息内容

		//1 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式不正确，请使用\"to|张三|你好呀\"格式。\n")
			return
		}
		//2 根据用户名 得到对方user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("该用户名不存在\n")
			return
		}
		//3 获取消息内容，通过对方的User对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无效内容，请使用\"to|张三|你好呀\"格式。\n")
		}
		remoteUser.SendMsg(this.Name + "对您说：" + content)

	} else if len(msg) > 12 && msg[:12] == "createGroup|" {
		//消息格式：createGroup|group1
		groupName := strings.Split(msg, "|")[1]
		if groupName == "" {
			this.SendMsg("消息格式不正确，请使用\"createGroup|group1\"格式。\n")
			return
		}
		NewGroup(groupName).AddGroupUser(this)
		this.SendMsg("创建群聊成功\n")
	} else if len(msg) > 10 && msg[:10] == "joinGroup|" {
		//消息格式：joinGroup|group1|张三
		groupName := strings.Split(msg, "|")[1]
		if groupName == "" {
			this.SendMsg("消息格式不正确，请使用\"joinGroup|group1|张三\"格式。\n")
			return
		}
		remoteName := strings.Split(msg, "|")[2]
		if remoteName == "" {
			this.SendMsg("消息格式不正确，请使用\"joinGroup|group1|张三\"格式。\n")
			return
		}
		//2 根据用户名 得到对方user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("该用户名不存在\n")
			return
		}
		for _, group := range remoteUser.Groups {
			if group.GroupName == groupName {
				this.SendMsg("用户已经加入该群聊\n")
				return
			}
		}
		for i := 0; i < len(this.Groups); i++ {
			if this.Groups[i].GroupName == groupName {
				this.Groups[i].AddGroupUser(remoteUser)
				this.SendMsg("加入群聊成功\n")
				return
			}
		}
		this.SendMsg("该群聊不存在\n")

	} else if msg == "showGroup" {
		for range this.Groups {
			this.SendMsg("群聊名称：" + this.Groups[0].GroupName + "\n")
		}

	} else if len(msg) > 10 && msg[:10] == "groupChat|" {
		//消息格式：groupChat|group1|你好呀
		groupName := strings.Split(msg, "|")[1]
		if groupName == "" {
			this.SendMsg("消息格式不正确，请使用\"groupChat|group1|你好呀\"格式。\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无效内容，请使用\"groupChat|group1|你好呀\"格式。\n")
		}
		for _, group := range this.Groups {
			if group.GroupName == groupName {
				for _, user := range group.GroupUsers {
					user.SendMsg(this.Name + "对群聊" + groupName + "说：" + content)
				}
				return
			}
		}
		this.SendMsg("该群聊不存在\n")

	} else {
		this.server.BroadCast(this, msg)
	}
}

type Group struct {
	GroupUsers []*User
	GroupName  string
	num        int
}

func NewGroup(groupName string) *Group {
	group := &Group{
		GroupName: groupName,
	}
	return group
}

func (this *Group) AddGroupUser(user *User) {
	this.GroupUsers = append(this.GroupUsers, user)
	this.num++
	user.Groups = append(user.Groups, this)
}
