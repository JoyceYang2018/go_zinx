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

// Test PreHandler
func (this *PingRouter) PreHandler(request ziface.IRequest){
	fmt.Println("Call Router PreHandler..")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("Before ping...\n"))
	if err != nil{
		fmt.Println("call back before ping error")
	}
}

// Test Handler
func (this *PingRouter) Handler(request ziface.IRequest){
	fmt.Println("Call Router Handler..")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("ping...ping...ping...\n"))
	if err != nil{
		fmt.Println("call back ping...ping...ping... error")
	}
}

// Test PostHandler
func (this *PingRouter) PostHandler(request ziface.IRequest){
	fmt.Println("Call Router PostHandler..")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("After ping...\n"))
	if err != nil{
		fmt.Println("call back after ping error")
	}
}


func main()  {
	//1 创建爱你一个server句柄 使用Zinx的api
	s := znet.NewServer("[zinx V0.3]")

	//2 给当前zinx框架添加一个自定义router
	s.AddRouter(&PingRouter{})

	//3 启动server
	s.Serve()
}
