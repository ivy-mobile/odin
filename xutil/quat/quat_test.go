package quat

import (
	"math"
	"testing"

	"github.com/ivy-mobile/odin/xutil/vec"
)

func almost(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9
}

func quatAlmost(a, b Quat) bool {
	return a.AlmostEqual(b, 1e-9)
}

func vecAlmost(a, b vec.Vec3) bool {
	return a.AlmostEqual(b, 1e-9)
}

func TestIdentityRotateVec3(t *testing.T) {
	v := vec.Vec3{X: 1, Y: 2, Z: 3}
	got := Identity().RotateVec3(v)
	if !vecAlmost(got, v) {
		t.Fatalf("RotateVec3 = %#v, want %#v", got, v)
	}
}

func TestFromAxisAngleRotateVec3(t *testing.T) {
	q := FromAxisAngle(vec.Vec3{Z: 1}, math.Pi/2)
	got := q.RotateVec3(vec.Vec3{X: 1})
	want := vec.Vec3{Y: 1}
	if !vecAlmost(got, want) {
		t.Fatalf("RotateVec3 = %#v, want %#v", got, want)
	}
}

func TestLookRotationIdentity(t *testing.T) {
	got := LookRotation(vec.Vec3{Z: 1})
	if !quatAlmost(got, Identity()) {
		t.Fatalf("LookRotation(+Z) = %#v, want identity", got)
	}
}

func TestLookRotationRotateForward(t *testing.T) {
	tests := []vec.Vec3{
		{X: 1},
		{Y: 1},
		{Z: -2},
		{X: 1, Y: 2, Z: 3},
	}

	for _, forward := range tests {
		q := LookRotation(forward)
		got := q.RotateVec3(vec.Vec3{Z: 1})
		want := forward.Normalize()
		if !vecAlmost(got, want) {
			t.Fatalf("LookRotation(%#v) rotate +Z = %#v, want %#v", forward, got, want)
		}
	}
}

func TestLookRotationZeroForward(t *testing.T) {
	got := LookRotation(vec.Vec3{})
	if !quatAlmost(got, Identity()) {
		t.Fatalf("LookRotation(zero) = %#v, want identity", got)
	}
}

func TestNormalizeZero(t *testing.T) {
	got := (Quat{}).Normalize()
	if !got.IsZero() {
		t.Fatalf("Normalize zero = %#v, want zero", got)
	}
}

func TestConjugateAndInverse(t *testing.T) {
	q := Quat{X: 1, Y: 2, Z: 3, W: 4}

	if got := q.Conjugate(); got != (Quat{X: -1, Y: -2, Z: -3, W: 4}) {
		t.Fatalf("Conjugate = %#v", got)
	}

	got := q.MulQuat(q.Inverse())
	if !quatAlmost(got, Identity()) {
		t.Fatalf("q * q^-1 = %#v, want identity", got)
	}
}

func TestMulQuatComposeRotation(t *testing.T) {
	qz := FromAxisAngle(vec.Vec3{Z: 1}, math.Pi/2)
	qy := FromAxisAngle(vec.Vec3{Y: 1}, math.Pi/2)

	composed := qy.MulQuat(qz)
	v := vec.Vec3{X: 1}

	got := composed.RotateVec3(v)
	want := qy.RotateVec3(qz.RotateVec3(v))
	if !vecAlmost(got, want) {
		t.Fatalf("composed rotate = %#v, want %#v", got, want)
	}
}

func TestLerpAndNlerp(t *testing.T) {
	a := Identity()
	b := Quat{W: 2}

	if got := a.Lerp(b, 0.25); !quatAlmost(got, Quat{W: 1.25}) {
		t.Fatalf("Lerp = %#v", got)
	}

	n := a.Nlerp(FromAxisAngle(vec.Vec3{Z: 1}, math.Pi), 0.5)
	if !almost(n.Len(), 1) {
		t.Fatalf("Nlerp len = %v, want 1", n.Len())
	}
}

func TestSlerp(t *testing.T) {
	a := Identity()
	b := FromAxisAngle(vec.Vec3{Z: 1}, math.Pi)

	half := a.Slerp(b, 0.5)
	got := half.RotateVec3(vec.Vec3{X: 1})
	want := vec.Vec3{Y: 1}
	if !vecAlmost(got, want) {
		t.Fatalf("Slerp rotate = %#v, want %#v", got, want)
	}
}

func TestSlerp1(t *testing.T) {
	v1 := Quat{1, 2, 3, 4}
	v2 := Quat{5, 6, 7, 8}

	t.Logf("1:%v", v1.Slerp(v2, 0.5))
	t.Logf("2:%v", v1.SlerpX(v2, 0.5))
}

func BenchmarkLen2(b *testing.B) {
	v1 := Quat{1, 2, 3, 4}

	b.Run("Len2", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = v1.Len2()
		}
	})

	b.Run("Len", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = v1.Len()
		}
	})
}
