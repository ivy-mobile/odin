package member_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ivy-mobile/odin/envelope"
	"github.com/ivy-mobile/odin/member"
)

func TestBase(t *testing.T) {
	lastActiveTime := time.Now()
	m := member.NewBase(
		1001,
		true,
		"player",
		"avatar",
		"unknown",
		true,
		lastActiveTime,
		true,
		3,
		2001,
	)

	assert.Equal(t, int64(1001), m.ID())
	assert.True(t, m.IsRobot())
	assert.Equal(t, "player", m.Nickname())
	assert.Equal(t, "avatar", m.Avatar())
	assert.Equal(t, "unknown", m.Gender())
	assert.True(t, m.IsReady())
	assert.Equal(t, lastActiveTime, m.LastActiveTime())
	assert.True(t, m.Offline())
	assert.Equal(t, 3, m.SeatID())
	assert.Equal(t, 2001, m.RoomID())

	snapshot := m.Snapshot()
	assert.Equal(t, int64(1001), snapshot.GetUid())
	assert.Equal(t, "player", snapshot.GetNickname())
	assert.Equal(t, "avatar", snapshot.GetAvatar())
	assert.Equal(t, "unknown", snapshot.GetGender())
	assert.Equal(t, int32(3), snapshot.GetSeatId())
	assert.True(t, snapshot.GetIsReady())
	assert.NotNil(t, snapshot.GetMeta())
}

type gameMember struct {
	*member.Base
	score int64
}

func (m *gameMember) Snapshot() *envelope.Member {
	snapshot := m.Base.Snapshot()
	snapshot.Meta["score"] = envelope.Int64V(m.score)
	return snapshot
}

func TestSpaceWithExtendedMember(t *testing.T) {
	space := member.NewSpace[*gameMember]()
	first := &gameMember{Base: newTestBase(1001), score: 10}
	second := &gameMember{Base: newTestBase(1002), score: 20}

	space.Set(first)
	space.Set(second)
	assert.Equal(t, 2, space.Count())

	got, ok := space.Get(first.ID())
	require.True(t, ok)
	assert.Same(t, first, got)
	assert.Equal(t, int64(10), got.score)
	assert.Equal(t, int64(10), got.Snapshot().GetMeta()["score"].GetI64())

	replacement := &gameMember{Base: newTestBase(1001), score: 30}
	space.Set(replacement)
	assert.Equal(t, 2, space.Count(), "replacing a member must not increase the count")

	ids := make(map[int64]struct{})
	space.Range(func(id int64, _ *gameMember) {
		ids[id] = struct{}{}
	})
	assert.Equal(t, map[int64]struct{}{1001: {}, 1002: {}}, ids)
	assert.Len(t, space.All(), 2)

	space.Remove(9999)
	assert.Equal(t, 2, space.Count(), "removing a missing member must not decrease the count")
	space.Remove(1001)
	assert.Equal(t, 1, space.Count())
	_, ok = space.Get(1001)
	assert.False(t, ok)
}

func newTestBase(id int64) *member.Base {
	return member.NewBase(id, false, "", "", "", false, time.Time{}, false, 0, 0)
}

var _ member.Member = (*member.Base)(nil)
var _ member.Member = (*gameMember)(nil)
