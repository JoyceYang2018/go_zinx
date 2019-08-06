package ziface

//路由抽象接口， 路由里的数据都是IRequest

type IRouter interface {
	//在处理conn业务之前的钩子方法hook
	PreHandler(request IRequest)
	//在处理conn业务的主方法hook
	Handler(request IRequest)
	//在处理conn业务之后的钩子方法hook
	PostHandler(request IRequest)
}