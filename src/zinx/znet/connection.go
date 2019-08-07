package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"zinx/ziface"
)

/*
	链接模块
*/

type Connection struct {
	//当前链接的socket TCP套接字
	Conn *net.TCPConn

	//链接的ID
	ConnID uint32

	//当前的链接状态
	isClose bool

	//告知当前链接已经退出/停止的 channel
	ExitChan chan bool

	//消息的管理MsgId和对应的处理业务API关系
	MsgHandler ziface.IMsgHandler
}

//初始化链接模块的方法
func NewCoinnection(conn *net.TCPConn, connID uint32, handler ziface.IMsgHandler) *Connection{
	c := &Connection{
		Conn: conn,
		ConnID: connID,
		MsgHandler: handler,
		isClose:false,
		ExitChan: make(chan bool, 1),
	}
	return c
}

// 链接的读业务方法
func (c *Connection) StartReader(){
	fmt.Println("Reader Coroutine is running....")
	defer fmt.Println("connID=", c.ConnID, "Reader is exit, remote addr is ", c.RemoteAddr().String())
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

		//从路由中找到注册绑定的conn对应的router调用
		// 根据绑定好的MsgId 找到对应处理api业务 执行
		go c.MsgHandler.DoMsgHandler(&req)
	}
}

//启动链接 让当前的链接准备开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start() ... ConnID = ", c.ConnID)
	//启动从当前链接的读数据的业务
	go c.StartReader()
	//TODO 启动从当前链接写数据的业务
}
//停止链接 结束当前链接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop().. CoonID=", c.ConnID)

	//如果当前链接已经关闭
	if c.isClose == true{
		return
	}
	c.isClose = true

	//关闭socket链接
	c.Conn.Close()

	//回收资源
	close(c.ExitChan)
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
	if _,err := c.Conn.Write(binaryMsg); err != nil{
		fmt.Println("Write msg id", msgId, "error:", err)
		return errors.New("conn Write err")
	}

	return nil
}