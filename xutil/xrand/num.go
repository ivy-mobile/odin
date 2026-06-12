package xrand

import (
	cryptorand "crypto/rand"
	"math/big"
)

// Int [min, max) 随机数生成
// crypto/rand 生成安全的随机数,相比math/rand性能更好，推荐使用
func Int(minValue, maxValue int) int {
	if minValue >= maxValue || maxValue == 0 {
		return maxValue
	}
	result, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(maxValue-minValue)))
	return int(result.Int64()) + minValue
}

// Int64 [min, max) 随机数生成
// crypto/rand 生成安全的随机数,相比math/rand性能更好，推荐使用
func Int64(minValue, maxValue int64) int64 {
	if minValue >= maxValue || maxValue == 0 {
		return maxValue
	}
	result, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(maxValue-minValue))
	return result.Int64() + minValue
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
func Float32(minValue, maxValue float32) float32 {
	if minValue == maxValue {
		return minValue
	}

	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}

	return minValue + globalRand.Float32()*(maxValue-minValue)
}

// Float64 生成[min,max)范围间的64位浮点数
func Float64(minValue, maxValue float64) float64 {
	if minValue == maxValue {
		return minValue
	}

	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}

	return minValue + globalRand.Float64()*(maxValue-minValue)
}
