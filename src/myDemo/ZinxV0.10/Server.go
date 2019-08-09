package main

import (
	"fmt"
	"zinx/ziface"
	"zinx/znet"
)

//基于Zinx框架来开发的， 服务器端应用程序

//ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

//hello zinx test 自定义路由
type HelloZinxRouter struct {
	znet.BaseRouter
}

// Test Handler
func (this *PingRouter) Handler(request ziface.IRequest){
	fmt.Println("Call PingRouter Handler..")
	// 先读取客户端数据 再回写ping..ping..ping
	fmt.Println("recv from client, msgid:", request.GetMsgId(), "data:", string(request.GetData()))
	err := request.GetConnection().SendMsg(200, []byte("ping..ping..ping"))
	if err != nil {
		fmt.Println(err)
	}
}

// Hello Handler
func (this *HelloZinxRouter) Handler(request ziface.IRequest){
	fmt.Println("Call HelloZinxRouter Handler..")
	// 先读取客户端数据 再回写ping..ping..ping
	fmt.Println("recv from client, msgid:", request.GetMsgId(), "data:", string(request.GetData()))
	err := request.GetConnection().SendMsg(201, []byte("Hello Welcome to Zinx"))
	if err != nil {
		fmt.Println(err)
	}
}

// 创建链接之后执行钩子函数
func DoConnBegin(conn ziface.IConnection){
	fmt.Println("=======> DoConnbegin is Called...")
	if err := conn.SendMsg(202, []byte("DoConnection Begin")); err != nil{
		fmt.Println(err)
	}

	// 给当前链接设置一些属性
	fmt.Println("Set conn Name, Home...")
	conn.SetProperty("Name", "yazi")
	conn.SetProperty("Home", "https://www.baidu.com")
}

// 链接断开之前执行钩子函数
func DoConnLost(conn ziface.IConnection){
	fmt.Println("=======> DoConnLost is Called...")
	fmt.Println("connId=", conn.GetConnID(), "is called...")

	// 获取链接属性
	if name, err := conn.GetProperty("Name"); err == nil{
		fmt.Println("Name=", name)
	}
	if home, err := conn.GetProperty("Home"); err == nil{
		fmt.Println("Home=", home)
	}
}


func main()  {
	//1 创建爱你一个server句柄 使用Zinx的api
	s := znet.NewServer("[zinx V0.5]")

	//2 给当前Server注册链接Hook钩子函数
	s.SetOnConnStart(DoConnBegin)
	s.SetOnConnStop(DoConnLost)

	//3 给当前zinx框架添加自定义router
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})

	//4 启动server
	s.Serve()
}
