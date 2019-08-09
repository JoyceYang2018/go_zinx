package znet

import (
	"errors"
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

//iServer的接口实现， 定义一个Server的服务器模块
type Server struct {
	//服务器名称
	Name string
	//服务器绑定的ip版本
	IPVersion string
	//服务器监听的ip
	IP string
	//服务器监听的端口
	Port int
	//当前的server的消息管理模块，用来绑定msgId和对应的处理业务API关系
	MsgHandler ziface.IMsgHandler
	// 该server的链接管理器
	ConnMgr ziface.IConnManager
	// 该Server创建链接之后自动调用的Hook函数--OnConnStart
	OnConnStart func(conn ziface.IConnection)
	// 该Server销毁链接之后自动调用的Hook函数--OnConnStop
	OnConnStop func(conn ziface.IConnection)

}

// 定义当前客户端链接的所绑定的handle api（目前是写死的， 以后优化由用户自定义）
func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
	//回显业务
	fmt.Println("[Conn Handle] CallBackToClient...")
	if _, err := conn.Write(data[:cnt]); err!=nil{
		fmt.Println("write back buf err", err)
		return errors.New("CallBackToClient error")
	}
	return nil
}

// 启动服务器
func (s *Server) Start(){
	fmt.Printf("[Zinx] Server Name:%s, listener at IP:%s, Port:%d is starting\n",
		utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	fmt.Printf("[Zinx] Version %s, MaxConn:%d, MaxPackageSize:%d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPackageSize)

	go func() {
		//0 开启消息队列及worker工作池
		s.MsgHandler.StartWorkerPool()

		//1 获取一个TCP的addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error : ", err)
			return
		}

		//2 监听服务器的地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err ", err)
			return
		}
		fmt.Println("start Zinx server succ, ", s.Name, "succ, Listening ..")
		var cid uint32
		cid = 0


		//3 阻塞的等待客户端的链接 处理客户端链接业务（读写）
		for {
			//如果有客户端链接过来， 阻塞会返回
			conn, err := listenner.AcceptTCP()
			if err != nil{
				fmt.Println("Accept err", err)
				continue
			}

			// 设置最大链接个数的判断 如果超过最大链接数量 则关闭此链接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn{
				// TODO 给客户端响应一个超出最大连接的错误包
				conn.Close()
				continue
			}

			//将处理新连接的业务方法和conn进行绑定 得到我们的链接模块
			dealConn := NewCoinnection(s, conn, cid, s.MsgHandler)
			cid ++

			//启动当前的链接业务处理
			go dealConn.Start()

		}
	}()
}

// 停止服务器
func (s *Server) Stop(){
	//将一些服务器资源，状态或一些开辟的链接信息 进行停止或者回收
	fmt.Println("[STOP] Zinx server name ", s.Name)
	s.ConnMgr.ClearConn()
}

// 运行服务器
func (s *Server) Serve(){
	//启动derver的服务功能
	s.Start()

	// TODO 做一些启动服务器之后的额外业务

	//阻塞状态
	select{}
}

//路由功能：给当前的服务注册一个路由方法， 供客户端的链接处理使用
func (s *Server) AddRouter(msgId uint32, router ziface.IRouter){
	s.MsgHandler.AddRouter(msgId, router)
	fmt.Println("Add Router Succ!")
}

func (s *Server) GetConnMgr() ziface.IConnManager{
	return s.ConnMgr
}

//初始化Server模块的方法
func NewServer(name string) ziface.IServer{
	s := &Server{
		Name: utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP: utils.GlobalObject.Host,
		Port: utils.GlobalObject.TcpPort,
		MsgHandler: NewMsgHandler(),
		ConnMgr: NewConnManager(),
	}
	return s
}

// 注册OnConnStart钩子函数的方法
func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnection)){
	s.OnConnStart = hookFunc
}

// 注册OnConnStop钩子函数的方法
func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnection)){
	s.OnConnStop = hookFunc
}

// 调用OnConnStart钩子函数的方法
func (s *Server) CallOnConnStart(conn ziface.IConnection){
	if s.OnConnStart != nil{
		fmt.Println("-----> Call OnConnStart() ...")
		s.OnConnStart(conn)
	}
}

// 调用OnConnStop钩子函数的方法
func (s *Server) CallOnConnStop(conn ziface.IConnection){
	if s.OnConnStop != nil{
		fmt.Println("-----> Call OnConnStop() ...")
		s.OnConnStop(conn)
	}
}