package znet

import (
	"errors"
	"fmt"
	"sync"
	"zinx/ziface"
)

// 链接管理模块
type ConnManager struct {
	connections map[uint32] ziface.IConnection // 管理的链接集合
	connLock sync.RWMutex // 保护链接集合的读写锁
}

// 创建当前链接的方法
func NewConnManager() *ConnManager{
	return &ConnManager{
		connections: make(map[uint32] ziface.IConnection),
	}
}


// 添加链接
func (cm *ConnManager) Add(conn ziface.IConnection){
	// 保护共享资源map，加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	// 将conn加入到cm中
	cm.connections[conn.GetConnID()] = conn
	fmt.Println("connId=", conn.GetConnID(), " add to ConnManage successfully: conn num=", cm.Len())
}
// 删除链接
func (cm *ConnManager) Remove(conn ziface.IConnection){
	// 保护共享资源map，加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	// 删除链接信息
	delete(cm.connections, conn.GetConnID())
	fmt.Println("connId=", conn.GetConnID(), " remove from ConnManage successfully: conn num=", cm.Len())
}

// 根据connId获取链接
func (cm *ConnManager) Get(connId uint32) (ziface.IConnection, error){
	// 保护共享资源map，加读锁
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	if conn, ok := cm.connections[connId]; ok{
		// 找到了
		return conn, nil
	}else{
		return nil, errors.New("connection not FOUND")
	}
}
// 得到当前链接总数
func (cm *ConnManager) Len() int{
	return len(cm.connections)
}
// 清除并种植所有的链接
func (cm *ConnManager) ClearConn(){
	// 保护共享资源map，加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	// 删除conn病停止conn的工作
	for connId, conn:= range cm.connections{
		//停止
		conn.Stop()
		// 删除
		delete(cm.connections, connId)
	}

	fmt.Println("Clear All connections secc! conn num=", cm.Len())
}