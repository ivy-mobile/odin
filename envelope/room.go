package envelope

import (
	"time"
)

// NewRoomConfig 房间配置
func NewRoomConfig(
	maxPlayerNum,
	minPlayerNum,
	roomType,
	mode,
	gameId int,
	extra map[string]*Value) *RoomConfig {

	return &RoomConfig{
		MaxPlayerNum: uint32(maxPlayerNum),
		RoomType:     int32(roomType),
		Mode:         int32(mode),
		GameId:       int32(gameId),
		MinPlayerNum: uint32(minPlayerNum),
		Extra:        extra,
	}
}

// NewRoomInfo 房间信息
func NewRoomInfo(
	roomId int64,
	roomName string,
	config *RoomConfig,
	players []*Player,
	createTime, updateTime time.Time,
	state RoomInfo_State,
	extra map[string]*Value) *RoomInfo {
	return &RoomInfo{
		RoomId:     roomId,
		RoomName:   roomName,
		Config:     config,
		Players:    players,
		CreateTime: createTime.UnixMilli(),
		UpdateTime: updateTime.UnixMilli(),
		State:      state,
		Extra:      extra,
	}
}
