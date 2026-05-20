package quat

import (
	"math"

	"github.com/ivy-mobile/odin/xutil/vec"
)

const defaultNormalizeEps = 1e-9

// Quat 四元数(Quaternion)，用于表示旋转
//
// 约定：
//   - (X, Y, Z) 是虚部向量
//   - W 是实部
//   - 单位四元数通常表示旋转
type Quat struct {
	X, Y, Z, W float64
}

// Identity 返回单位四元数
func Identity() Quat {
	return Quat{W: 1}
}

// LookRotation 根据朝向向量创建旋转四元数，语义对齐 Unity 的 Quaternion.LookRotation(forward)
//
// 约定：
//   - 结果会把局部前向轴 +Z 旋转到 forward 方向
//   - 世界上方向默认使用 +Y
//   - 若 forward 为零向量，返回单位四元数
func LookRotation(forward vec.Vec3) Quat {
	return LookRotationUp(forward, vec.Vec3{Y: 1})
}

// LookRotationX 从方向向量创建四元数（LookRotation）
// 模拟 Unity 的 Quaternion.LookRotation(forward)
func LookRotationX(forward vec.Vec3) Quat {
	// 归一化前向向量
	forward = forward.Normalize()
	if forward.Len() < 0.0001 {
		return Identity()
	}

	// 使用世界上方向作为参考
	up := vec.Vec3{0, 1, 0}

	// 计算右向量
	right := up.Cross(forward).Normalize()
	if right.Len() < 0.0001 {
		// forward 与 up 平行，使用备用上方向
		right = vec.Vec3{1, 0, 0}
	}

	// 重新计算上向量
	up = forward.Cross(right)

	// 从旋转矩阵构建四元数
	// 旋转矩阵: [right, up, forward]
	m00, m01, m02 := right.X, up.X, forward.X
	m10, m11, m12 := right.Y, up.Y, forward.Y
	m20, m21, m22 := right.Z, up.Z, forward.Z

	trace := m00 + m11 + m22
	var q Quat

	if trace > 0 {
		s := math.Sqrt(trace+1.0) * 2 // s = 4 * qw
		q.W = 0.25 * s
		q.X = (m21 - m12) / s
		q.Y = (m02 - m20) / s
		q.Z = (m10 - m01) / s
	} else if m00 > m11 && m00 > m22 {
		s := math.Sqrt(1.0+m00-m11-m22) * 2 // s = 4 * qx
		q.W = (m21 - m12) / s
		q.X = 0.25 * s
		q.Y = (m01 + m10) / s
		q.Z = (m02 + m20) / s
	} else if m11 > m22 {
		s := math.Sqrt(1.0+m11-m00-m22) * 2 // s = 4 * qy
		q.W = (m02 - m20) / s
		q.X = (m01 + m10) / s
		q.Y = 0.25 * s
		q.Z = (m12 + m21) / s
	} else {
		s := math.Sqrt(1.0+m22-m00-m11) * 2 // s = 4 * qz
		q.W = (m10 - m01) / s
		q.X = (m02 + m20) / s
		q.Y = (m12 + m21) / s
		q.Z = 0.25 * s
	}

	return q.Normalize()
}

// LookRotationUp 根据朝向向量和上方向创建旋转四元数
//
// 若 forward 为零向量，返回单位四元数。
// 若 upwards 无效或与 forward 平行，会自动选择一个与 forward 正交的备用上方向。
func LookRotationUp(forward, upwards vec.Vec3) Quat {
	f := forward.Normalize()
	if f.IsZero() {
		return Identity()
	}

	up := upwards.Normalize()
	if up.IsZero() {
		up = fallbackUp(f)
	}

	right := up.Cross(f).Normalize()
	if right.IsZero() {
		up = fallbackUp(f)
		right = up.Cross(f).Normalize()
		if right.IsZero() {
			return Identity()
		}
	}

	up = f.Cross(right).Normalize()
	return quatFromBasis(right, up, f)
}

// FromAxisAngle 根据旋转轴和弧度创建四元数
//
// 若 axis 为零向量，返回单位四元数
func FromAxisAngle(axis vec.Vec3, rad float64) Quat {
	unit := axis.Normalize()
	if unit.IsZero() {
		return Identity()
	}
	half := rad * 0.5
	s := math.Sin(half)
	return Quat{
		X: unit.X * s,
		Y: unit.Y * s,
		Z: unit.Z * s,
		W: math.Cos(half),
	}
}

// Add 返回四元数加法 a + b
func (a Quat) Add(b Quat) Quat {
	return Quat{a.X + b.X, a.Y + b.Y, a.Z + b.Z, a.W + b.W}
}

// Sub 返回四元数减法 a - b
func (a Quat) Sub(b Quat) Quat {
	return Quat{a.X - b.X, a.Y - b.Y, a.Z - b.Z, a.W - b.W}
}

// Mul 返回四元数与标量的乘法 a * s
func (a Quat) Mul(s float64) Quat {
	return Quat{a.X * s, a.Y * s, a.Z * s, a.W * s}
}

// Div 返回四元数与标量的除法 a / s
func (a Quat) Div(s float64) Quat {
	return Quat{a.X / s, a.Y / s, a.Z / s, a.W / s}
}

// Neg 返回相反四元数 -a
func (a Quat) Neg() Quat {
	return Quat{-a.X, -a.Y, -a.Z, -a.W}
}

// Dot 返回点积
func (a Quat) Dot(b Quat) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z + a.W*b.W
}

// Len2 返回长度平方
func (a Quat) Len2() float64 {
	return a.Dot(a)
}

// Len 返回长度
func (a Quat) Len() float64 {
	return math.Sqrt(a.Len2())
}

// IsZero 返回是否为精确零四元数。
func (a Quat) IsZero() bool {
	return a.X == 0 && a.Y == 0 && a.Z == 0 && a.W == 0
}

// NearZero 返回是否接近零四元数
func (a Quat) NearZero(eps float64) bool {
	if eps <= 0 {
		eps = defaultNormalizeEps
	}
	return a.Len2() <= eps*eps
}

// Normalize 返回单位长度四元数
func (a Quat) Normalize() Quat {
	return a.NormalizeEps(defaultNormalizeEps)
}

// NormalizeEps 返回单位长度四元数，允许自定义零阈值
func (a Quat) NormalizeEps(eps float64) Quat {
	if eps <= 0 {
		eps = defaultNormalizeEps
	}
	l2 := a.Len2()
	if l2 <= eps*eps {
		return Quat{}
	}
	return a.Mul(1 / math.Sqrt(l2))
}

// Conjugate 返回共轭四元数
func (a Quat) Conjugate() Quat {
	return Quat{-a.X, -a.Y, -a.Z, a.W}
}

// Inverse 返回逆四元数
//
// 若 a 为零四元数，返回零四元数
func (a Quat) Inverse() Quat {
	l2 := a.Len2()
	if l2 == 0 {
		return Quat{}
	}
	return a.Conjugate().Div(l2)
}

// MulQuat 返回 Hamilton 乘积 a * b
//
// 对旋转来说，结果表示先应用 b，再应用 a
func (a Quat) MulQuat(b Quat) Quat {
	return Quat{
		X: a.W*b.X + a.X*b.W + a.Y*b.Z - a.Z*b.Y,
		Y: a.W*b.Y - a.X*b.Z + a.Y*b.W + a.Z*b.X,
		Z: a.W*b.Z + a.X*b.Y - a.Y*b.X + a.Z*b.W,
		W: a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z,
	}
}

// RotateVec3 用四元数旋转三维向量
//
// 若四元数接近零，返回原向量
func (a Quat) RotateVec3(v vec.Vec3) vec.Vec3 {
	q := a.Normalize()
	if q.IsZero() {
		return v
	}
	p := Quat{X: v.X, Y: v.Y, Z: v.Z}
	r := q.MulQuat(p).MulQuat(q.Conjugate())
	return vec.Vec3{X: r.X, Y: r.Y, Z: r.Z}
}

// Lerp 返回线性插值
func (a Quat) Lerp(b Quat, t float64) Quat {
	return a.Add(b.Sub(a).Mul(t))
}

// Nlerp 返回归一化线性插值
//
// 若点积为负，会翻转 b 以走最短弧
func (a Quat) Nlerp(b Quat, t float64) Quat {
	if a.Dot(b) < 0 {
		b = b.Neg()
	}
	return a.Lerp(b, t).Normalize()
}

// Slerp 返回球面线性插值
//
// 若两者非常接近，会退化为 Nlerp 以避免数值不稳定
func (a Quat) Slerp(b Quat, t float64) Quat {
	qa := a.Normalize()
	qb := b.Normalize()
	if qa.IsZero() {
		return qb
	}
	if qb.IsZero() {
		return qa
	}

	cosTheta := qa.Dot(qb)
	if cosTheta < 0 {
		qb = qb.Neg()
		cosTheta = -cosTheta
	}

	if cosTheta > 0.9995 {
		return qa.Lerp(qb, t).Normalize()
	}

	if cosTheta > 1 {
		cosTheta = 1
	}
	if cosTheta < -1 {
		cosTheta = -1
	}

	theta := math.Acos(cosTheta)
	sinTheta := math.Sin(theta)
	if math.Abs(sinTheta) <= defaultNormalizeEps {
		return qa
	}

	w1 := math.Sin((1-t)*theta) / sinTheta
	w2 := math.Sin(t*theta) / sinTheta
	return qa.Mul(w1).Add(qb.Mul(w2)).Normalize()
}

// SlerpX 球面线性插值（模拟 Unity 的 Quaternion.Slerp）
func (q Quat) SlerpX(target Quat, t float64) Quat {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// 计算点积
	dot := q.Dot(target)

	// 如果点积为负，取反一个四元数以获得最短路径
	if dot < 0 {
		target = Quat{-target.X, -target.Y, -target.Z, -target.W}
		dot = -dot
	}

	// 如果四元数非常接近，使用线性插值避免除零
	if dot > 0.9995 {
		return Quat{
			X: q.X + (target.X-q.X)*t,
			Y: q.Y + (target.Y-q.Y)*t,
			Z: q.Z + (target.Z-q.Z)*t,
			W: q.W + (target.W-q.W)*t,
		}.Normalize()
	}

	// 球面插值
	theta0 := math.Acos(dot)
	theta := theta0 * t
	sinTheta := math.Sin(theta)
	sinTheta0 := math.Sin(theta0)

	s0 := math.Cos(theta) - dot*sinTheta/sinTheta0
	s1 := sinTheta / sinTheta0

	return Quat{
		X: q.X*s0 + target.X*s1,
		Y: q.Y*s0 + target.Y*s1,
		Z: q.Z*s0 + target.Z*s1,
		W: q.W*s0 + target.W*s1,
	}.Normalize()
}

// AlmostEqual 判断两个四元数是否在 eps 内近似相等
func (a Quat) AlmostEqual(b Quat, eps float64) bool {
	return a.Sub(b).NearZero(eps)
}

func fallbackUp(forward vec.Vec3) vec.Vec3 {
	if math.Abs(forward.Y) < 0.999 {
		return vec.Vec3{Y: 1}
	}
	return vec.Vec3{X: 1}
}

func quatFromBasis(right, up, forward vec.Vec3) Quat {
	m00, m01, m02 := right.X, up.X, forward.X
	m10, m11, m12 := right.Y, up.Y, forward.Y
	m20, m21, m22 := right.Z, up.Z, forward.Z

	trace := m00 + m11 + m22
	if trace > 0 {
		s := math.Sqrt(trace+1.0) * 2
		return Quat{
			X: (m21 - m12) / s,
			Y: (m02 - m20) / s,
			Z: (m10 - m01) / s,
			W: 0.25 * s,
		}.Normalize()
	}

	if m00 > m11 && m00 > m22 {
		s := math.Sqrt(1.0+m00-m11-m22) * 2
		return Quat{
			X: 0.25 * s,
			Y: (m01 + m10) / s,
			Z: (m02 + m20) / s,
			W: (m21 - m12) / s,
		}.Normalize()
	}

	if m11 > m22 {
		s := math.Sqrt(1.0+m11-m00-m22) * 2
		return Quat{
			X: (m01 + m10) / s,
			Y: 0.25 * s,
			Z: (m12 + m21) / s,
			W: (m02 - m20) / s,
		}.Normalize()
	}

	s := math.Sqrt(1.0+m22-m00-m11) * 2
	return Quat{
		X: (m02 + m20) / s,
		Y: (m12 + m21) / s,
		Z: 0.25 * s,
		W: (m10 - m01) / s,
	}.Normalize()
}
