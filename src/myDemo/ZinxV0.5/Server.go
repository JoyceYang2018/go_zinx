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

// Test Handler
func (this *PingRouter) Handler(request ziface.IRequest){
	fmt.Println("Call Router Handler..")
	// 先读取客户端数据 再回写ping..ping..ping
	fmt.Println("recv from client, msgid:", request.GetMsgId(), "data:", string(request.GetData()))
	err := request.GetConnection().SendMsg(1, []byte("ping..ping..ping"))
	if err != nil {
		fmt.Println(err)
	}
}


func main()  {
	//1 创建爱你一个server句柄 使用Zinx的api
	s := znet.NewServer("[zinx V0.5]")

	//2 给当前zinx框架添加一个自定义router
	s.AddRouter(&PingRouter{})

	//3 启动server
	s.Serve()
}
