package xrand

import (
	cryptorand "crypto/rand"
	"math/big"
)

// Int [min, max) 随机数生成
// crypto/rand 生成安全的随机数,相比math/rand性能更好，推荐使用
func Int(min, max int) int {

	if min >= max || max == 0 {
		return max
	}
	result, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max-min)))
	return int(result.Int64()) + min
}

// Int64 [min, max) 随机数生成
// crypto/rand 生成安全的随机数,相比math/rand性能更好，推荐使用
func Int64(min, max int64) int64 {

	if min >= max || max == 0 {
		return max
	}
	result, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max-min)))
	return result.Int64() + min
}

// IsBingo 是否中奖
//
//	percent: 概率 (%)
func IsBingo(percent int) bool {
	return Int(0, 10000) < percent*100
}

// IsBingo64 是否中奖
//
//	percent: 概率 (%)
func IsBingo64(percent int64) bool {
	return Int64(0, 10000) < percent*100
}

// Float32 生成[min,max)范围间的32位浮点数
func Float32(min, max float32) float32 {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + globalRand.Float32()*(max-min)
}

// Float64 生成[min,max)范围间的64位浮点数
func Float64(min, max float64) float64 {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + globalRand.Float64()*(max-min)
}
