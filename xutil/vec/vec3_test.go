package vec

import (
	"math"
	"testing"
)

func almost(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9
}

func vecAlmost(a, b Vec3) bool {
	return a.AlmostEqual(b, 1e-9)
}

func TestVec3NormalizeZero(t *testing.T) {
	got := (Vec3{}).Normalize()
	if !got.IsZero() {
		t.Fatalf("Normalize zero = %#v, want zero", got)
	}
}

func TestVec3Cross(t *testing.T) {
	x := Vec3{1, 0, 0}
	y := Vec3{0, 1, 0}
	got := x.Cross(y)
	want := Vec3{0, 0, 1}
	if !vecAlmost(got, want) {
		t.Fatalf("Cross = %#v, want %#v", got, want)
	}
}

func TestVec3ProjectOn(t *testing.T) {
	v := Vec3{2, 2, 0}
	axis := Vec3{1, 0, 0}
	got := v.ProjectOn(axis)
	want := Vec3{2, 0, 0}
	if !vecAlmost(got, want) {
		t.Fatalf("ProjectOn = %#v, want %#v", got, want)
	}
}

func TestVec3ReflectNonUnitNormal(t *testing.T) {
	v := Vec3{1, -2, 0}
	n := Vec3{0, 2, 0}
	got := v.Reflect(n)
	want := Vec3{1, 2, 0}
	if !vecAlmost(got, want) {
		t.Fatalf("Reflect = %#v, want %#v", got, want)
	}
}

func TestVec3RejectAndPlaneProjection(t *testing.T) {
	v := Vec3{3, 4, 5}
	n := Vec3{0, 2, 0} // 非单位法线，等价于 Y 轴法线

	reject := v.RejectFrom(n)
	wantReject := Vec3{3, 0, 5}
	if !vecAlmost(reject, wantReject) {
		t.Fatalf("RejectFrom = %#v, want %#v", reject, wantReject)
	}

	projPlane := v.ProjectOnPlane(n)
	if !vecAlmost(projPlane, wantReject) {
		t.Fatalf("ProjectOnPlane = %#v, want %#v", projPlane, wantReject)
	}

	slide := v.SlideOnPlane(n)
	if !vecAlmost(slide, wantReject) {
		t.Fatalf("SlideOnPlane = %#v, want %#v", slide, wantReject)
	}

	unitN := Vec3{0, 1, 0}
	projPlaneUnit := v.ProjectOnPlaneUnit(unitN)
	if !vecAlmost(projPlaneUnit, wantReject) {
		t.Fatalf("ProjectOnPlaneUnit = %#v, want %#v", projPlaneUnit, wantReject)
	}

	slideUnit := v.SlideOnPlaneUnit(unitN)
	if !vecAlmost(slideUnit, wantReject) {
		t.Fatalf("SlideOnPlaneUnit = %#v, want %#v", slideUnit, wantReject)
	}
}

func TestVec3ClampLen(t *testing.T) {
	v := Vec3{3, 4, 0}
	got := v.ClampLen(2)
	if !almost(got.Len(), 2) {
		t.Fatalf("ClampLen len = %v, want 2", got.Len())
	}

	if keep := v.ClampLen(10); !vecAlmost(keep, v) {
		t.Fatalf("ClampLen should keep vector, got %#v want %#v", keep, v)
	}
}

func TestVec3Bounce(t *testing.T) {
	v := Vec3{3, -4, 0}
	n := Vec3{0, 2, 0} // 非单位法线

	if got := v.Bounce(n, 1); !vecAlmost(got, Vec3{3, 4, 0}) {
		t.Fatalf("Bounce e=1 = %#v, want %#v", got, Vec3{3, 4, 0})
	}

	if got := v.Bounce(n, 0); !vecAlmost(got, Vec3{3, 0, 0}) {
		t.Fatalf("Bounce e=0 = %#v, want %#v", got, Vec3{3, 0, 0})
	}

	if got := v.Bounce(n, 0.5); !vecAlmost(got, Vec3{3, 2, 0}) {
		t.Fatalf("Bounce e=0.5 = %#v, want %#v", got, Vec3{3, 2, 0})
	}
}

func TestVec3Lerp(t *testing.T) {
	a := Vec3{0, 0, 0}
	b := Vec3{10, 20, 30}
	got := a.Lerp(b, 0.25)
	want := Vec3{2.5, 5, 7.5}
	if !vecAlmost(got, want) {
		t.Fatalf("Lerp = %#v, want %#v", got, want)
	}
}

func TestVec3MoveToward(t *testing.T) {
	a := Vec3{0, 0, 0}
	b := Vec3{10, 0, 0}

	if got := a.MoveToward(b, 3); !vecAlmost(got, Vec3{3, 0, 0}) {
		t.Fatalf("MoveToward step = %#v, want %#v", got, Vec3{3, 0, 0})
	}

	if got := a.MoveToward(b, 20); !vecAlmost(got, b) {
		t.Fatalf("MoveToward overshoot clamp = %#v, want %#v", got, b)
	}

	if got := a.MoveToward(b, 0); !vecAlmost(got, a) {
		t.Fatalf("MoveToward zero delta = %#v, want %#v", got, a)
	}
}

func TestMoveTowards(t *testing.T) {
	current := Vec3{0, 0, 0}
	target := Vec3{10, 0, 0}

	if got := MoveTowards(current, target, 3); !vecAlmost(got, Vec3{3, 0, 0}) {
		t.Fatalf("MoveTowards step = %#v, want %#v", got, Vec3{3, 0, 0})
	}

	if got := MoveTowards(current, target, 20); !vecAlmost(got, target) {
		t.Fatalf("MoveTowards overshoot clamp = %#v, want %#v", got, target)
	}

	if got := MoveTowards(current, target, -3); !vecAlmost(got, Vec3{-3, 0, 0}) {
		t.Fatalf("MoveTowards negative delta = %#v, want %#v", got, Vec3{-3, 0, 0})
	}

	if got := MoveTowards(target, target, 3); !vecAlmost(got, target) {
		t.Fatalf("MoveTowards same point = %#v, want %#v", got, target)
	}
}

func TestVec3AngleAndCos(t *testing.T) {
	x := Vec3{1, 0, 0}
	y := Vec3{0, 1, 0}
	if !almost(x.CosAngle(y), 0) {
		t.Fatalf("CosAngle = %v, want 0", x.CosAngle(y))
	}
	if !almost(x.Angle(y), math.Pi/2) {
		t.Fatalf("Angle = %v, want %v", x.Angle(y), math.Pi/2)
	}
}

func TestVec3AlmostEqual(t *testing.T) {
	a := Vec3{1, 2, 3}
	b := Vec3{1 + 1e-10, 2 - 1e-10, 3}
	if !a.AlmostEqual(b, 1e-9) {
		t.Fatalf("AlmostEqual should be true")
	}
	if a.AlmostEqual(Vec3{1.1, 2, 3}, 1e-9) {
		t.Fatalf("AlmostEqual should be false")
	}
}
