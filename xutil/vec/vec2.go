package vec

import "math"

// defaultVec2Eps 是 Vec2 归一化/近零判断的默认阈值
const defaultVec2Eps = 1e-9

// Vec2 表示二维向量（也可用于二维点坐标）
//
// 常见 2D 游戏语义：
//   - 位置：屏幕坐标/世界坐标（X,Y）
//   - 方向：朝向向量
//   - 速度/加速度：每个轴的速度分量
type Vec2 struct {
	X, Y float64
}

// Add 返回向量加法 a + b（分量相加）
func (a Vec2) Add(b Vec2) Vec2 {
	return Vec2{a.X + b.X, a.Y + b.Y}
}

// Sub 返回向量减法 a - b（分量相减）
//
// 几何意义：
//   - a.Sub(b) 是“从 b 指向 a”的向量
func (a Vec2) Sub(b Vec2) Vec2 {
	return Vec2{a.X - b.X, a.Y - b.Y}
}

// Mul 返回向量与标量的乘法 a * s（整体缩放）
func (a Vec2) Mul(s float64) Vec2 {
	return Vec2{a.X * s, a.Y * s}
}

// Div 返回向量与标量的除法 a / s
//
// 注意：
//   - s == 0 会产生 Inf/NaN（Go 浮点规则）
func (a Vec2) Div(s float64) Vec2 {
	return Vec2{a.X / s, a.Y / s}
}

// Hadamard 返回逐分量乘法（component-wise multiply）
//
// 用途：
//   - 按轴缩放
//   - 屏蔽某个轴，例如 v.Hadamard(Vec2{1,0})
func (a Vec2) Hadamard(b Vec2) Vec2 {
	return Vec2{a.X * b.X, a.Y * b.Y}
}

// Neg 返回相反向量 -a
func (a Vec2) Neg() Vec2 {
	return Vec2{-a.X, -a.Y}
}

// Dot 返回点积（内积）a · b
//
// 若 a、b 都是单位向量，则结果为夹角余弦 cos(θ)。
func (a Vec2) Dot(b Vec2) float64 {
	return a.X*b.X + a.Y*b.Y
}

// Cross 返回二维叉积的 z 分量（标量）
//
// 数学含义：
//   - cross2(a,b) = ax*by - ay*bx
//
// 几何意义：
//   - > 0 表示 b 在 a 的“逆时针侧”
//   - < 0 表示 b 在 a 的“顺时针侧”
//   - = 0 表示共线（或接近共线）
//
// 游戏用途：
//   - 判定左右转向
//   - 计算 2D 面积符号
func (a Vec2) Cross(b Vec2) float64 {
	return a.X*b.Y - a.Y*b.X
}

// Len2 返回长度平方 |a|^2
func (a Vec2) Len2() float64 {
	return a.Dot(a)
}

// Len 返回欧几里得长度 |a|
func (a Vec2) Len() float64 {
	return math.Sqrt(a.Len2())
}

// IsZero 返回是否为精确零向量
func (a Vec2) IsZero() bool {
	return a.X == 0 && a.Y == 0
}

// NearZero 返回向量是否接近零。
//
// eps <= 0 时使用默认阈值 1e-9
func (a Vec2) NearZero(eps float64) bool {
	if eps <= 0 {
		eps = defaultVec2Eps
	}
	return a.Len2() <= eps*eps
}

// Normalize 返回单位向量（长度约为 1），零向量时返回零向量
func (a Vec2) Normalize() Vec2 {
	return a.NormalizeEps(defaultVec2Eps)
}

// NormalizeEps 与 Normalize 类似，但允许自定义阈值
//
// eps <= 0 时使用默认阈值 1e-9
func (a Vec2) NormalizeEps(eps float64) Vec2 {
	if eps <= 0 {
		eps = defaultVec2Eps
	}
	l2 := a.Len2()
	if l2 <= eps*eps {
		return Vec2{}
	}
	return a.Mul(1 / math.Sqrt(l2))
}

// PerpCCW 返回逆时针旋转 90 度后的向量
//
// (x,y) -> (-y,x)
func (a Vec2) PerpCCW() Vec2 {
	return Vec2{-a.Y, a.X}
}

// PerpCW 返回顺时针旋转 90 度后的向量
//
// (x,y) -> (y,-x)
func (a Vec2) PerpCW() Vec2 {
	return Vec2{a.Y, -a.X}
}

// Rotate 返回向量绕原点旋转 rad（弧度）后的结果
//
// 约定：
//   - rad > 0 为逆时针旋转
func (a Vec2) Rotate(rad float64) Vec2 {
	c := math.Cos(rad)
	s := math.Sin(rad)
	return Vec2{
		X: a.X*c - a.Y*s,
		Y: a.X*s + a.Y*c,
	}
}

// Lerp 返回线性插值：a*(1-t) + b*t
func (a Vec2) Lerp(b Vec2, t float64) Vec2 {
	return a.Add(b.Sub(a).Mul(t))
}

// ClampLen 把向量长度限制到不超过 maxLen
//
// maxLen <= 0 时返回零向量
func (a Vec2) ClampLen(maxLen float64) Vec2 {
	if maxLen <= 0 {
		return Vec2{}
	}
	l2 := a.Len2()
	max2 := maxLen * maxLen
	if l2 <= max2 {
		return a
	}
	if l2 == 0 {
		return Vec2{}
	}
	return a.Mul(maxLen / math.Sqrt(l2))
}

// ProjectOn 返回 a 在 b 上的投影向量
//
// 若 b 为零向量，返回零向量
func (a Vec2) ProjectOn(b Vec2) Vec2 {
	denom := b.Len2()
	if denom == 0 {
		return Vec2{}
	}
	return b.Mul(a.Dot(b) / denom)
}

// RejectFrom 返回 a 相对 b 的垂直分量（a - proj_b(a)）
func (a Vec2) RejectFrom(b Vec2) Vec2 {
	return a.Sub(a.ProjectOn(b))
}

// Reflect 返回向量 a 关于法线 n 的反射向量
//
// n 不要求单位化；若 n 为零向量，返回原向量
func (a Vec2) Reflect(n Vec2) Vec2 {
	denom := n.Len2()
	if denom == 0 {
		return a
	}
	return a.Sub(n.Mul(2 * a.Dot(n) / denom))
}

// Bounce 返回带恢复系数的反弹结果
//
// restitution：
//   - 1：完全弹性
//   - 0：不反弹，只保留切线分量
//   - <0 会被钳制为 0
func (a Vec2) Bounce(n Vec2, restitution float64) Vec2 {
	if restitution < 0 {
		restitution = 0
	}
	denom := n.Len2()
	if denom == 0 {
		return a
	}
	vn := n.Mul(a.Dot(n) / denom)
	vt := a.Sub(vn)
	return vt.Sub(vn.Mul(restitution))
}

// CosAngle 返回 a 与 b 的夹角余弦值
//
// 若任一向量接近零向量，返回 0
func (a Vec2) CosAngle(b Vec2) float64 {
	la2 := a.Len2()
	lb2 := b.Len2()
	if la2 <= defaultVec2Eps*defaultVec2Eps || lb2 <= defaultVec2Eps*defaultVec2Eps {
		return 0
	}
	c := a.Dot(b) / math.Sqrt(la2*lb2)
	if c > 1 {
		return 1
	}
	if c < -1 {
		return -1
	}
	return c
}

// Angle 返回 a 与 b 的夹角（弧度，范围 [0, π]）
func (a Vec2) Angle(b Vec2) float64 {
	return math.Acos(a.CosAngle(b))
}

// SignedAngle 返回从 a 旋转到 b 的有符号夹角（弧度，范围约 [-π, π]）
//
// 正值表示逆时针，负值表示顺时针
func (a Vec2) SignedAngle(b Vec2) float64 {
	if a.NearZero(defaultVec2Eps) || b.NearZero(defaultVec2Eps) {
		return 0
	}
	return math.Atan2(a.Cross(b), a.Dot(b))
}

// AlmostEqual 判断两个向量是否在 eps 阈值内近似相等
//
// eps <= 0 时使用默认阈值 1e-9
func (a Vec2) AlmostEqual(b Vec2, eps float64) bool {
	return a.Sub(b).NearZero(eps)
}

// Distance2 返回两点间距离平方
func (a Vec2) Distance2(b Vec2) float64 {
	return a.Sub(b).Len2()
}

// Distance 返回两点间实际距离
func (a Vec2) Distance(b Vec2) float64 {
	return math.Sqrt(a.Distance2(b))
}

// MoveToward 返回从 a 朝 b 移动最多 maxDelta 后的位置/向量
//
// 若 maxDelta 足够大，不会越过目标，直接返回 b
func (a Vec2) MoveToward(b Vec2, maxDelta float64) Vec2 {
	if maxDelta <= 0 {
		return a
	}
	delta := b.Sub(a)
	dist2 := delta.Len2()
	if dist2 == 0 || dist2 <= maxDelta*maxDelta {
		return b
	}
	return a.Add(delta.Mul(maxDelta / math.Sqrt(dist2)))
}
