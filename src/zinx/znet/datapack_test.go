package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

// 只负责测试datapack的拆包 封包的单元测试
func TestDataPack(t *testing.T) {
	// 模拟的服务器
	//1 创建socketTCP
	listener, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil{
		fmt.Println("server listen err", err)
		return
	}

	// 创建一个go 承载 负责从客户端处理业务
	go func() {
		//2 从客户端读取数据 拆包处理
		for {
			conn, err := listener.Accept()
			if err != nil{
				fmt.Println("server accept error", err)
			}

			go func(conn net.Conn) {
				// 处理客户端请求
				//---------> 拆包过程
				// 定义一个拆包的对象dp
				dp := NewDataPack()
				for {
					//1 第一次从conn读 把包的head读出来
					headData := make([]byte, dp.GetHeadLen())
					_, err := io.ReadFull(conn, headData)
					if err != nil{
						fmt.Println("read head error", err)
						return
					}

					msgHead, err := dp.Unpack(headData)
					if err != nil{
						fmt.Println("server unpack err", err)
						return
					}

					if msgHead.GetMsgLen() > 0 {
						//msg是有数据的， 需要进行第二次读取
						//2 第二次从conn读 根据head的datalen 再读取出data
						//msg := msgHead.(*Message)
						//msg.Data =
					}

				}

			}(conn)
		}
	}()



	// 模拟客户端
}