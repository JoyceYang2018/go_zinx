package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/utils"
	"zinx/ziface"
)

/*
	链接模块
*/

type Connection struct {
	// 当前conn隶属于哪个server
	TcpServer ziface.IServer

	//当前链接的socket TCP套接字
	Conn *net.TCPConn

	//链接的ID
	ConnID uint32

	//当前的链接状态
	isClose bool

	//告知当前链接已经退出/停止的 channel (由Reader告知Writer退出)
	ExitChan chan bool

	//用于读缓冲的管道， 用于读写goroutine之间的消息通信
	msgChan chan []byte

	//消息的管理MsgId和对应的处理业务API关系
	MsgHandler ziface.IMsgHandler

	// 链接属性集合
	property map[string]interface{}

	// 保护链接属性的锁
	propertyLock sync.RWMutex
}

//初始化链接模块的方法
func NewCoinnection(server ziface.IServer, conn *net.TCPConn, connID uint32, handler ziface.IMsgHandler) *Connection{
	c := &Connection{
		TcpServer: server,
		Conn: conn,
		ConnID: connID,
		MsgHandler: handler,
		isClose:false,
		msgChan: make(chan []byte),
		ExitChan: make(chan bool, 1),
		property: make(map[string]interface{}),
	}

	// 将conn加入到ConnManager中
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

// 链接的读业务方法
func (c *Connection) StartReader(){
	fmt.Println("[Reader Coroutine is running...]")
	defer fmt.Println("connID=", c.ConnID, "[Reader is exit!] remote addr is ", c.RemoteAddr().String())
	defer c.Stop()

	for{
		//读取客户端的数据到buf中
		//buf := make([]byte, utils.GlobalObject.MaxPackageSize)
		//_, err := c.Conn.Read(buf)
		//if err != nil{
		//	fmt.Println("recv buf err", err)
		//	continue
		//}
		// 创建一个拆包 解包的对象
		dp := NewDataPack()

		// 读取客户端的msghead 二进制流 8字节
		headData := make([]byte, dp.GetHeadLen())
		if _,err := io.ReadFull(c.GetTCPConnection(), headData); err !=nil{
			fmt.Println("read msg head err", err)
			break
		}

		// 拆包，得到MsgId和MsgDataLen 放在msg消息中
		msg, err := dp.Unpack(headData)
		if err != nil{
			fmt.Println("unpack err", err)
			break
		}
		// 根据dataLen 再次读取data 放在msg.Data中
		var data []byte
		if msg.GetMsgLen() > 0{
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil{
				fmt.Println("read msg data err", err)
				break
			}
		}
		msg.SetData(data)
		//得到当前链接数据的Request请求数据
		req := Request{
			conn:c,
			msg:msg,
		}

		// 判断是否开启了工作池模式
		if utils.GlobalObject.WorkerPoolSize >0{
			//已经开启了工作池模式，将消息发送给worker工作池处理即可
			c.MsgHandler.SendMsgToTaskQueue(&req)
		}else{
			//从路由中找到注册绑定的conn对应的router调用
			// 根据绑定好的MsgId 找到对应处理api业务 执行
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

// 写消息的goroutine 专门发送给客户端消息的模块
func (c *Connection) StartWriter(){
	fmt.Println("[Writer Goroutine is running...]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	// 不断的阻塞等待channel的消息， 进行写给客户端
	for {
		select {
		case data := <- c.msgChan:
			//有数据写给客户端
			if _, err := c.Conn.Write(data); err != nil{
				fmt.Println("Send data err", err)
				return
			}
		case <-c.ExitChan:
			// 代表Reader已经退出， 此时Writer也要退出
			return
		}
	}
}

//启动链接 让当前的链接准备开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start() ... ConnID = ", c.ConnID)
	//启动从当前链接的读数据的业务
	go c.StartReader()
	//启动从当前链接写数据的业务
	go c.StartWriter()

	// 按照开发者传递进来的 创建链接之后需要调用的处理业务 执行对应的hook方法
	c.TcpServer.CallOnConnStart(c)
}

//停止链接 结束当前链接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop().. CoonID=", c.ConnID)

	//如果当前链接已经关闭
	if c.isClose == true{
		return
	}
	c.isClose = true

	// 调用开发者注册的 销毁链接之前 需要执行的业务Hook函数
	c.TcpServer.CallOnConnStop(c)

	//关闭socket链接
	c.Conn.Close()

	// 告知Writer关闭
	c.ExitChan <- true

	// 将当前链接从connmgr中删除
	c.TcpServer.GetConnMgr().Remove(c)

	//回收资源
	close(c.ExitChan)
	close(c.msgChan)
}
//获取当前链接的绑定socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}
//获取当前链接模块的链接id
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}
//获取远程客户端的TCP状态 IPport
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 提供sendmsg方法 将要发送给客户端的数据 先进行封包 再发送
func (c * Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClose == true {
		return errors.New("Connection closed when send msg")
	}

	// 将data进行封包 MsgDataLen/MsgId/Data
	dp := NewDataPack()
	binaryMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil{
		fmt.Println("Pack err msg id:", msgId)
		return errors.New("Pack err msg")
	}

	// 将数据写回给客户端
	c.msgChan <- binaryMsg

	return nil
}

// 设置链接属性
func(c *Connection) SetProperty(key string, value interface{}){
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	// 添加一个链接属性
	c.property[key] = value
}

// 获取链接属性
func(c *Connection) GetProperty(key string) (interface{}, error){
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	// 读取属性
	if value, ok := c.property[key];ok{
		return value, nil
	}else{
		return nil, errors.New("no property found")
	}
}

// 移除链接属性
func(c *Connection) RemoveProperty(key string){
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	// 删除属性
	delete(c.property, key)
}