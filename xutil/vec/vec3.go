package vec

import "math"

// defaultNormalizeEps 是向量归一化/近零判断的默认阈值
//
// 说明：
//   - 浮点计算里“理论上的 0”常常会变成一个非常小的数（例如 1e-15）
//   - 这里使用 1e-9 作为工程上的默认阈值，兼顾稳定性与常见游戏尺度
//   - 如果你的世界单位很大或很小，可以改用 NormalizeEps / NearZero 传入自定义阈值
const defaultNormalizeEps = 1e-9

// Vec3 表示三维向量（也可用于三维点坐标）
//
// 约定：
//   - 当它表示“位置”时，(X, Y, Z) 是世界坐标
//   - 当它表示“方向/速度/加速度”时，(X, Y, Z) 是对应分量
//   - 本类型不区分“点”和“向量”，由调用语义决定
type Vec3 struct {
	X, Y, Z float64
}

func Zero3() Vec3 {
	return Vec3{}
}

// MoveTowards 返回从 current 平滑移动到 target 的结果
// 模拟 Unity 的 Vector3.MoveTowards
func MoveTowards(current, target Vec3, maxDelta float64) Vec3 {
	diff := target.Sub(current)
	magnitude := diff.Len()

	if magnitude <= maxDelta || magnitude < 0.00001 {
		return target
	}
	return current.Add(diff.Normalize().Mul(maxDelta))
}

// Add 返回向量加法 a + b（分量相加）
//
// 数学含义： a + b = (ax + bx, ay + by, az + bz)
//
// 游戏/工程用途：
//   - 位置积分：pos = pos.Add(vel.Mul(dt))（每 tick 用速度推进位置）
//   - 位移叠加：把多个偏移量累计起来
//
// 注意：
//   - 这是“分量相加”，不是“长度相加”
func (a Vec3) Add(b Vec3) Vec3 {
	return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

// Sub 返回向量减法 a - b（分量相减）
//
// 数学含义： a - b = (ax - bx, ay - by, az - bz)
//
// 几何意义：
//   - a.Sub(b) 得到的是“从 b 指向 a 的向量”（b->a）
//     例如：toTarget = targetPos.Sub(myPos) 表示“我指向目标”的方向向量（未归一化）
//
// 游戏/工程用途：
//   - 求目标方向：dir = targetPos.Sub(curPos)
//   - 求位移差：delta = end.Sub(start)
//
// 注意：
//   - 顺序很重要：target.Sub(cur) 与 cur.Sub(target) 方向相反
func (a Vec3) Sub(b Vec3) Vec3 {
	return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

// Mul 返回向量与标量的乘法 a * s（把向量整体缩放）
//
// 数学含义： a * s = (ax*s, ay*s, az*s)
//
// 几何意义：
//   - s > 1：放大向量长度
//   - 0 < s < 1：缩小向量长度
//   - s < 0：长度缩放同时方向翻转
//
// 游戏/工程用途：
//   - 速度乘 dt 得到位移：delta = vel.Mul(dt)
//   - 归一化时用倒数缩放：unit = v.Mul(1/|v|)
//   - 施加加速度：vel = vel.Add(acc.Mul(dt))（如果你这么设计）
//
// 注意：
//   - 这是“标量乘法”，不是叉乘（cross product）
func (a Vec3) Mul(s float64) Vec3 {
	return Vec3{a.X * s, a.Y * s, a.Z * s}
}

// Hadamard 返回逐分量乘法（也叫 component-wise multiply）
//
// 数学含义： (ax,ay,az) ⊙ (bx,by,bz) = (ax*bx, ay*by, az*bz)
//
// 游戏/工程用途：
//   - 非均匀缩放：按轴缩放尺寸、速度、偏移
//   - 遮罩系数：例如只保留水平速度 v.Hadamard(Vec3{1,0,1})
//
// 注意：
//   - 这不是点积，也不是叉积。
func (a Vec3) Hadamard(b Vec3) Vec3 {
	return Vec3{a.X * b.X, a.Y * b.Y, a.Z * b.Z}
}

// Div 返回向量与标量的除法 a / s
//
// 数学含义： a / s = (ax/s, ay/s, az/s)
//
// 游戏/工程用途：
//   - 已知总时间时，反推平均速度：vel = delta.Div(dt)
//   - 已知质量时，做简化缩放（例如某些数值变换）
//
// 注意：
//   - s == 0 会产生 Inf/NaN（与 Go 浮点规则一致），调用方应自行保证安全
//   - 如果是做归一化，优先用 Normalize/NormalizeEps（带零向量保护）
func (a Vec3) Div(s float64) Vec3 {
	return Vec3{a.X / s, a.Y / s, a.Z / s}
}

// Neg 返回相反向量 -a
//
// 数学含义： -(ax,ay,az) = (-ax,-ay,-az)
//
// 游戏/工程用途：
//   - 反向速度/反向力
//   - 快速翻转方向（等价于 Mul(-1)）
func (a Vec3) Neg() Vec3 {
	return Vec3{-a.X, -a.Y, -a.Z}
}

// Dot 返回点积（内积）a · b
//
// 数学含义： a · b = ax*bx + ay*by + az*bz
//
// 几何意义：
//
//	a · b = |a| * |b| * cos(θ)
//	- 如果 a、b 都是单位向量（normalize 后），点积就是 cos(θ)
//	  · 接近 1：方向几乎一致
//	  · 接近 0：接近垂直
//	  · 小于 0：方向相反（“在背后”）
//
// 游戏/工程用途：
//   - 判定朝向/夹角：dot(forward, toTarget) > 0 表示在前方
//   - 投影长度：projLen = v.Dot(unitDir)
//   - 碰撞/反射等几何计算的基础
//
// 注意：
//   - 如果没归一化，dot 不是 cos 值，别误用
func (a Vec3) Dot(b Vec3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// Len2 返回向量长度的平方（模长平方）|a|^2
//
// 数学含义： |a|^2 = a · a = ax^2 + ay^2 + az^2
//
// 游戏/工程用途：
//   - 高频逻辑里用来做“距离比较/速度比较”非常常用：
//     if v.Len2() < r*r { ... }
//     因为避免了 sqrt，性能更好。
//
// 注意：
//   - Len2 不是长度本身，而是长度的平方
//   - 比较时记得把阈值也平方（r*r）
func (a Vec3) Len2() float64 {
	return a.Dot(a)
}

// IsZero 返回是否为精确零向量（每个分量都恰好为 0）
//
// 注意：
//   - 这是“精确比较”，通常只适合初始化值、明确赋零后的检查
//   - 对浮点计算结果，通常应优先使用 NearZero
func (a Vec3) IsZero() bool {
	return a.X == 0 && a.Y == 0 && a.Z == 0
}

// NearZero 返回向量是否接近零（使用长度平方与 eps 比较）
//
// 数学含义：
//   - 判断 |a| <= eps（实现中用 |a|^2 <= eps^2，避免 sqrt）
//
// 游戏/工程用途：
//   - 判定速度是否“几乎静止”
//   - 判定法线/方向是否有效，避免后续归一化产生 NaN
//
// eps <= 0 时使用默认阈值 1e-9
func (a Vec3) NearZero(eps float64) bool {
	if eps <= 0 {
		eps = defaultNormalizeEps
	}
	return a.Len2() <= eps*eps
}

// Len 返回向量的欧几里得长度（模长）|a|
//
// 数学含义： |a| = sqrt(ax^2 + ay^2 + az^2)
//
// 游戏/工程用途：
//   - 从速度向量得到速度标量：speed = vel.Len()
//   - 归一化前需要真实长度：dir = v / |v|
//
// 性能说明：
//   - 内部调用 sqrt，比 Len2 慢，tick 热路径尽量少用
//   - 能用 Len2 比较就用 Len2
func (a Vec3) Len() float64 {
	return math.Sqrt(a.Len2())
}

// Normalize 返回 a 的单位向量（归一化方向向量）
//
// 数学含义： normalize(a) = a / |a|
//
// 几何意义：
//   - 返回的向量长度为 1（单位长度），只保留方向信息
//   - 常见用途是把“方向”和“速度大小”拆开：
//     dir = vel.Normalize()
//     vel = dir.Mul(speed)
//
// 游戏/工程用途：
//   - 追踪/转向：只需要方向，不希望被长度影响
//   - 计算夹角/点积：单位向量 dot 才能直接得到 cos
//   - 重构速度：Vel = Dir * Speed
//
// 零向量处理（非常重要）：
//   - 如果 |a| 很小（< 1e-9），直接返回零向量 Vec3{}，避免除 0 导致 NaN
//   - 这种情况常出现在：速度为 0、两个点重合、数值抖动等
//
// 注意：
//   - 调用方如果“必须要一个合法方向”，需要对返回零向量做兜底处理
func (a Vec3) Normalize() Vec3 {
	return a.NormalizeEps(defaultNormalizeEps)
}

// NormalizeEps 与 Normalize 类似，但允许自定义零向量阈值
//
// 游戏/工程用途：
//   - 不同系统使用不同尺度时（例如米制世界、像素世界）自定义容差
//   - 碰撞法线、速度方向等高频逻辑统一一个 epsilon
//
// eps <= 0 时使用默认阈值 1e-9
func (a Vec3) NormalizeEps(eps float64) Vec3 {
	if eps <= 0 {
		eps = defaultNormalizeEps
	}
	l2 := a.Len2()
	if l2 <= eps*eps {
		return Vec3{}
	}
	return a.Mul(1.0 / math.Sqrt(l2))
}

// Cross 返回叉积（向量垂直于 a 和 b）
//
// 数学含义： a × b = (ay*bz - az*by, az*bx - ax*bz, ax*by - ay*bx)
//
// 几何意义：
//   - 结果向量垂直于 a 和 b
//   - 叉积的长度 |a × b| 等于平行四边形的面积
//   - 叉积的方向用右手定则判断
//
// 游戏/工程用途：
//   - 计算法线
//   - 判定旋转方向
//   - 计算面积
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{
		a.Y*b.Z - a.Z*b.Y,
		a.Z*b.X - a.X*b.Z,
		a.X*b.Y - a.Y*b.X,
	}
}

// Lerp 返回线性插值：a*(1-t) + b*t
//
// 几何意义：
//   - t=0 返回 a
//   - t=1 返回 b
//   - 0<t<1 返回 a 到 b 之间的点
//
// 游戏/工程用途：
//   - 相机平滑、位置插值、颜色/参数插值（若复用此向量做参数容器）
//   - 网络同步中的简单插值
//
// t 常用范围是 [0,1]，但也支持外推（t < 0 或 t > 1）
func (a Vec3) Lerp(b Vec3, t float64) Vec3 {
	return a.Add(b.Sub(a).Mul(t))
}

// ClampLen 把向量长度限制到不超过 max
//
// 几何意义：
//   - 如果 |a| <= maxLen，保持不变
//   - 如果 |a| > maxLen，保持方向不变并缩短到 maxLen
//
// 游戏/工程用途：
//   - 限速（角色速度、抛射物速度）
//   - 限制输入叠加后的总移动向量长度，避免斜向移动更快
//
// maxLen <= 0 时返回零向量
func (a Vec3) ClampLen(maxLen float64) Vec3 {
	if maxLen <= 0 {
		return Vec3{}
	}
	l2 := a.Len2()
	max2 := maxLen * maxLen
	if l2 <= max2 {
		return a
	}
	if l2 == 0 {
		return Vec3{}
	}
	return a.Mul(maxLen / math.Sqrt(l2))
}

// ProjectOn 返回 a 在 b 上的投影向量
//
// 数学含义：
//   - proj_b(a) = ((a·b)/(b·b)) * b
//
// 几何意义：
//   - 返回 a 在 b 方向上的“平行分量”
//
// 游戏/工程用途：
//   - 求沿某方向的速度分量/位移分量
//   - 把速度拆分为“法线方向 + 切线方向”
//
// 如果 b 为零向量，返回零向量
func (a Vec3) ProjectOn(b Vec3) Vec3 {
	denom := b.Len2()
	if denom == 0 {
		return Vec3{}
	}
	return b.Mul(a.Dot(b) / denom)
}

// Reflect 返回向量 a 关于法线 n 的反射向量
//
// 数学含义（n 不必单位化）：
//   - r = a - 2 * proj_n(a)
//
// 游戏/工程用途：
//   - 子弹/射线/速度在墙面法线上的镜面反射
//   - 弹性碰撞的理想反射方向（不含能量损失）
//
// n 不要求是单位向量；若 n 为零向量，则返回原向量 a
func (a Vec3) Reflect(n Vec3) Vec3 {
	denom := n.Len2()
	if denom == 0 {
		return a
	}
	return a.Sub(n.Mul(2 * a.Dot(n) / denom))
}

// RejectFrom 返回 a 相对于 b 的“垂直分量”（也叫 rejection）
//
// 数学含义：
//   - reject_b(a) = a - proj_b(a)
//
// 游戏/工程用途：
//   - 从速度中剔除某方向分量（例如剔除法线方向，仅保留切线方向）
//   - 碰撞响应里分离“沿法线”和“沿接触面”的速度
func (a Vec3) RejectFrom(b Vec3) Vec3 {
	return a.Sub(a.ProjectOn(b))
}

// ProjectOnPlane 返回 a 在“法线为 n 的平面”上的投影向量
//
// 几何意义：
//   - 从 a 中移除法线方向分量，得到贴着平面的分量（切线分量）
//
// 游戏/工程用途：
//   - 地面移动：把输入方向投影到地面平面，避免爬坡/斜面时方向失真
//   - 碰撞后速度修正：去掉穿透法线的分量，保留沿表面滑动的分量
//
// 注意：
//   - 若 n 为零向量，返回原向量 a（无法定义平面法线）
func (a Vec3) ProjectOnPlane(n Vec3) Vec3 {
	return a.RejectFrom(n)
}

// ProjectOnPlaneUnit 返回 a 在“法线为单位向量 n”的平面上的投影向量
//
// 数学含义（要求 n 已归一化）：
//   - a_plane = a - (a·n) * n
//
// 游戏/工程用途：
//   - 碰撞系统已保证法线单位化时，可少一次 n·n 的除法
//   - 角色在地面/斜坡上的移动方向投影
//
// 注意：
//   - 若 n 不是单位向量，结果会不正确；不确定时请使用 ProjectOnPlane
func (a Vec3) ProjectOnPlaneUnit(n Vec3) Vec3 {
	return a.Sub(n.Mul(a.Dot(n)))
}

// SlideOnPlane 返回 a 在平面上的滑动向量（语义化别名）
//
// 与 ProjectOnPlane 的结果相同，但在碰撞代码里可读性更好：
//   - vel = vel.SlideOnPlane(hitNormal)
func (a Vec3) SlideOnPlane(n Vec3) Vec3 {
	return a.ProjectOnPlane(n)
}

// SlideOnPlaneUnit 返回 a 在“单位法线平面”上的滑动向量（语义化别名）
//
// 与 ProjectOnPlaneUnit 结果相同，适合碰撞法线已单位化的热路径：
//   - vel = vel.SlideOnPlaneUnit(hit.Normal)
func (a Vec3) SlideOnPlaneUnit(n Vec3) Vec3 {
	return a.ProjectOnPlaneUnit(n)
}

// Bounce 返回 a 关于法线 n 的反弹结果，并应用恢复系数 restitution
//
// 参数说明：
//   - restitution=1：理想弹性反弹（法线方向速度完全保留）
//   - restitution=0：法线方向速度完全消失（只保留切线/滑动分量）
//   - 0<restitution<1：常见阻尼反弹
//
// 数学含义：
//   - 将 a 拆为切线分量 vt 与法线分量 vn
//   - 反弹后：vt 保持，vn 反向并乘以 restitution
//
// 注意：
//   - n 不要求单位化；若 n 为零向量，返回原向量 a
//   - restitution < 0 会被钳制为 0（避免产生“向内吸附”的反常行为）
func (a Vec3) Bounce(n Vec3, restitution float64) Vec3 {
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

// MoveToward 返回从 a 朝 b 移动最多 maxDelta 后的位置/向量
//
// 几何意义：
//   - 若 a 到 b 的距离 <= maxDelta，直接返回 b（不会越过目标）
//   - 否则沿 a->b 方向前进 maxDelta
//
// 游戏/工程用途：
//   - 角色/相机平滑跟随（固定步长）
//   - 追踪导引、数值渐近，避免 Lerp 在变步长下的尾部拖慢
//
// 注意：
//   - maxDelta <= 0 时返回 a（不移动）
func (a Vec3) MoveToward(b Vec3, maxDelta float64) Vec3 {
	if maxDelta <= 0 {
		return a
	}
	delta := b.Sub(a)
	dist2 := delta.Len2()
	if dist2 == 0 {
		return b
	}
	if dist2 <= maxDelta*maxDelta {
		return b
	}
	return a.Add(delta.Mul(maxDelta / math.Sqrt(dist2)))
}

// CosAngle 返回 a 与 b 的夹角余弦值
//
// 数学含义：
//   - cos(θ) = (a·b) / (|a||b|)
//
// 游戏/工程用途：
//   - 视野判断（FOV）：dot(forward, dirToTarget) 与阈值比较
//   - 方向一致性判断、坡度角估算（结合法线）
//
// 数值稳定性：
//   - 内部会把结果钳制到 [-1,1]，避免浮点误差导致 acos 输入越界
//
// 若任一向量接近零向量，则返回 0（避免 NaN）
func (a Vec3) CosAngle(b Vec3) float64 {
	la2 := a.Len2()
	lb2 := b.Len2()
	if la2 <= defaultNormalizeEps*defaultNormalizeEps || lb2 <= defaultNormalizeEps*defaultNormalizeEps {
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

// Angle 返回 a 与 b 的夹角（弧度）
//
// 返回值范围：
//   - [0, π]
//
// 游戏/工程用途：
//   - 需要“真实角度”而不仅仅是大小比较时使用
//
// 性能说明：
//   - 内部调用 acos，比点积/余弦比较更慢；热路径优先使用 CosAngle 或 Dot
//
// 若任一向量接近零向量，返回 0
func (a Vec3) Angle(b Vec3) float64 {
	return math.Acos(a.CosAngle(b))
}

// AlmostEqual 判断两个向量是否在 eps 阈值内近似相等（按欧几里得距离）
//
// 数学含义：
//   - |a - b| <= eps
//
// 游戏/工程用途：
//   - 测试断言
//   - 防止因浮点误差导致状态机抖动（例如“是否到达目标点”）
//
// eps <= 0 时使用默认阈值 1e-9
func (a Vec3) AlmostEqual(b Vec3, eps float64) bool {
	return a.Sub(b).NearZero(eps)
}

// Distance2 返回两点间距离的平方
//
// 数学含义： |a - b|^2
//
// 游戏/工程用途：
//   - 距离比较（视野半径、触发范围、碰撞粗检）优先使用这个版本，避免 sqrt
func (a Vec3) Distance2(b Vec3) float64 {
	return a.Sub(b).Len2()
}

// Distance 返回两点之间的实际距离
//
// 数学含义： |a - b|
//
// 游戏/工程用途：
//   - 需要真实距离数值时使用（UI 显示、衰减计算等）
//
// 性能说明：
//   - 内部调用 sqrt；仅做阈值比较时优先使用 Distance2
func (a Vec3) Distance(b Vec3) float64 {
	return math.Sqrt(a.Distance2(b))
}
