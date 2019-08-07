package znet

import (
	"fmt"
	"strconv"
	"zinx/ziface"
)

// 消息处理模块的实现
type MsgHandler struct {
	// 存放每个MsgId对应的处理方法
	Apis map[uint32] ziface.IRouter
}

// 初始化创建MsgHandler方法
func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Apis: make(map[uint32] ziface.IRouter),
	}
}

// 调度执行对应的Router消息处理方法
func (mh *MsgHandler)DoMsgHandler(request ziface.IRequest){
	// 从Request中找到msgId
	handler, ok := mh.Apis[request.GetMsgId()]
	if !ok {
		fmt.Println("api msgId-", request.GetMsgId(), "is not found! need register!")
	}
	// 根据msgId调度对应的router业务
	handler.PreHandler(request)
	handler.Handler(request)
	handler.PostHandler(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandler)AddRouter(msgId uint32, router ziface.IRouter){
	//1 判断 当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgId];ok{
		// id已经注册了
		panic("repeat api, msgId="+strconv.Itoa(int(msgId)))
	}
	//2 添加msg与API绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api MshId=", msgId, "succ!")
}
