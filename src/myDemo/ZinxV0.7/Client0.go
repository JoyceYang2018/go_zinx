package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zinx/znet"
)

/*
	模拟客户端
*/
func main(){
	fmt.Println("client start ...")

	time.Sleep(1 * time.Second)

	//1 直接链接远程服务器， 得到一个conn链接
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err !=nil{
		fmt.Println("client start err, exit!")
		return
	}

	for{
		//2 发送封包的msg消息 MsgId:0
		dp := znet.NewDataPack()
		binaryMsg, err := dp.Pack(znet.NewMsgPackage(0, []byte("ZinxV0.6 client0 Test Message")))
		if err != nil{
			fmt.Println("Pack err", err)
			return
		}
		if _, err := conn.Write(binaryMsg); err != nil{
			fmt.Println("write err", err)
			return
		}

		// 服务器应该回回复一个message数据 MsfId:1 ping..ping..ping
		//1 先读取流中的Head 得到id和datalen
		binaryHead := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(conn, binaryHead); err != nil{
			fmt.Println("read head error", err)
			break
		}
		// 将二进制的head拆包到msg结构体
		msgHead, err := dp.Unpack(binaryHead)
		if err != nil{
			fmt.Println("client unpack msgHead err", err)
			break
		}

		if msgHead.GetMsgLen() > 0{
			//2 再根据datalen进行第二次读取 将Data读出来
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(conn, msg.Data); err != nil{
				fmt.Println("read msg data err", err)
				return
			}

			fmt.Println("------> Recv sercerMsg id:", msg.Id, "len:", msg.DataLen, "data", string(msg.Data))
		}

		//cpu阻塞
		time.Sleep(1*time.Second)
	}
}