package vec

import (
	"math"
	"testing"
)

func almost2(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9
}

func vec2Almost(a, b Vec2) bool {
	return a.AlmostEqual(b, 1e-9)
}

func TestVec2NormalizeZero(t *testing.T) {
	if got := (Vec2{}).Normalize(); !got.IsZero() {
		t.Fatalf("Normalize zero = %#v, want zero", got)
	}
}

func TestVec2CrossSign(t *testing.T) {
	x := Vec2{1, 0}
	y := Vec2{0, 1}
	if !almost2(x.Cross(y), 1) {
		t.Fatalf("x.Cross(y) = %v, want 1", x.Cross(y))
	}
	if !almost2(y.Cross(x), -1) {
		t.Fatalf("y.Cross(x) = %v, want -1", y.Cross(x))
	}
}

func TestVec2Perp(t *testing.T) {
	v := Vec2{2, 3}
	if got := v.PerpCCW(); !vec2Almost(got, Vec2{-3, 2}) {
		t.Fatalf("PerpCCW = %#v", got)
	}
	if got := v.PerpCW(); !vec2Almost(got, Vec2{3, -2}) {
		t.Fatalf("PerpCW = %#v", got)
	}
}

func TestVec2Rotate(t *testing.T) {
	v := Vec2{1, 0}
	got := v.Rotate(math.Pi / 2)
	if !vec2Almost(got, Vec2{0, 1}) {
		t.Fatalf("Rotate pi/2 = %#v, want (0,1)", got)
	}
}

func TestVec2ProjectReflectBounce(t *testing.T) {
	v := Vec2{3, -4}
	n := Vec2{0, 2} // 非单位法线

	if got := v.ProjectOn(n); !vec2Almost(got, Vec2{0, -4}) {
		t.Fatalf("ProjectOn = %#v, want %#v", got, Vec2{0, -4})
	}
	if got := v.RejectFrom(n); !vec2Almost(got, Vec2{3, 0}) {
		t.Fatalf("RejectFrom = %#v, want %#v", got, Vec2{3, 0})
	}
	if got := v.Reflect(n); !vec2Almost(got, Vec2{3, 4}) {
		t.Fatalf("Reflect = %#v, want %#v", got, Vec2{3, 4})
	}
	if got := v.Bounce(n, 0); !vec2Almost(got, Vec2{3, 0}) {
		t.Fatalf("Bounce e=0 = %#v, want %#v", got, Vec2{3, 0})
	}
	if got := v.Bounce(n, 0.5); !vec2Almost(got, Vec2{3, 2}) {
		t.Fatalf("Bounce e=0.5 = %#v, want %#v", got, Vec2{3, 2})
	}
}

func TestVec2Angles(t *testing.T) {
	x := Vec2{1, 0}
	y := Vec2{0, 1}
	if !almost2(x.CosAngle(y), 0) {
		t.Fatalf("CosAngle = %v, want 0", x.CosAngle(y))
	}
	if !almost2(x.Angle(y), math.Pi/2) {
		t.Fatalf("Angle = %v, want %v", x.Angle(y), math.Pi/2)
	}
	if !almost2(x.SignedAngle(y), math.Pi/2) {
		t.Fatalf("SignedAngle x->y = %v, want %v", x.SignedAngle(y), math.Pi/2)
	}
	if !almost2(y.SignedAngle(x), -math.Pi/2) {
		t.Fatalf("SignedAngle y->x = %v, want %v", y.SignedAngle(x), -math.Pi/2)
	}
}

func TestVec2MoveTowardAndClampLen(t *testing.T) {
	a := Vec2{0, 0}
	b := Vec2{10, 0}

	if got := a.MoveToward(b, 3); !vec2Almost(got, Vec2{3, 0}) {
		t.Fatalf("MoveToward = %#v, want %#v", got, Vec2{3, 0})
	}
	if got := a.MoveToward(b, 20); !vec2Almost(got, b) {
		t.Fatalf("MoveToward overshoot = %#v, want %#v", got, b)
	}

	v := Vec2{3, 4}
	if got := v.ClampLen(2); !almost2(got.Len(), 2) {
		t.Fatalf("ClampLen len = %v, want 2", got.Len())
	}
}
