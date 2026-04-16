package envelope

// NewPlayer 创建玩家信息
func NewPlayer(
	uid int64,
	nickname,
	avatar string,
	gender, seatId int,
	ready bool,
	extra map[string]*Value) *Player {

	return &Player{
		Uid:      uid,
		Nickname: nickname,
		Avatar:   avatar,
		Gender:   int32(gender),
		SeatId:   int32(seatId),
		IsReady:  ready,
		Extra:    extra,
	}
}
