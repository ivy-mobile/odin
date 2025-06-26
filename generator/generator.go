package generator

// RoomIdGenerator 房间ID生成器统一接口
type RoomIdGenerator interface {
	GenRoomID() (string, error)
}
