package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/ziface"
)

// 消息处理模块的实现
type MsgHandler struct {
	// 存放每个MsgId对应的处理方法
	Apis map[uint32] ziface.IRouter
	// 负责worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	// 业务工作worker池的数量
	WorkerPoolSize uint32

}

// 初始化创建MsgHandler方法
func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Apis: make(map[uint32] ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize, // 从全局配置中获取
		TaskQueue: make([]chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen),
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

// 启动一个worker工作池(开启工作池的动作 一个zinx框架只能发生一次)
func (mh *MsgHandler) StartWorkerPool(){
	// 根据workerpoolsize分别开启worker 每个worker用一个go来承载
	for i:=0; i<int(mh.WorkerPoolSize); i++{
		// 一个worker启动
		//1 给当前worker对应的channel消息队列 开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//2 启动当前worker 阻塞等待消息从channel传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 启动一个worker工作流程
func (mh *MsgHandler) StartOneWorker(workerId int, taskQueue chan ziface.IRequest){
	fmt.Println("WorkerId=", workerId, "is started...")

	// 不断的的阻塞等待对应的消息队列的消息
	for{
		select {
			//如果由于消息过来，出列的就是一个客户端的Request 执行当前Request所绑定的业务
			case request:= <- taskQueue:
				mh.DoMsgHandler(request)
		}
	}
}

// 将消息交给TaskQueue 由worker进行处理
func (mh *MsgHandler) SendMsgToTaskQueue(request ziface.IRequest) {
	//1 将消息平均分配给不同的Worker
	// 根据客户端建立的ConnId拉丝进行分配
	workerId:= request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnId=", request.GetConnection().GetConnID(),
		" request MsgId=", request.GetMsgId(),
		"to WorkerId=", workerId)

	//
	//2 将消息发送给对应的worker的TaskQueue即可
	mh.TaskQueue[workerId] <- request
}