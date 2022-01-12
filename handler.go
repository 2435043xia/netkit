package netkit

type PacketHandler[T any, P PacketType[T]] func(sess *Session[T, P], packet P)
type Middleware[T any, P PacketType[T]] func(sess *Session[T, P], packet P) error
type HandlerInfo[T any, P PacketType[T]] struct {
	MessageId int
	Handler   PacketHandler[T, P]
}

type Handler[T any, P PacketType[T]] interface {
	GetHandlerById(messageId int) *HandlerInfo[T,P]
	GetMiddlewares() []Middleware[T,P]
}
