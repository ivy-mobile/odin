package xrand

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	LetterSeed           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // 字母字
	LetterLowerSeed      = "abcdefghijklmnopqrstuvwxyz"                           // 小写字母
	LetterUpperSeed      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"                           // 大写字母
	DigitSeed            = "0123456789"                                           // 数字
	DigitWithoutZeroSeed = "123456789"                                            // 无0数字
	SymbolSeed           = "!\\\"#$%&'()*+,-./:;<=>?@[\\\\]^_`{|}~"               // 特殊字符
)

var globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// Str 生成指定长度的字符串
func Str(seed string, length int) (str string) {
	if length <= 0 {
		return ""
	}

	r := []rune(seed)
	n := len(r)
	if n == 0 {
		return
	}

	for i := 0; i < length; i++ {
		pos := globalRand.Intn(n)
		str += string(r[pos : pos+1])
	}

	return
}

// Letters 生成指定长度的字母字符串
func Letters(length int) string {
	return Str(LetterSeed, length)
}

// Digits 生成指定长度的数字字符串
func Digits(length int, hasLeadingZero ...bool) string {
	if len(hasLeadingZero) > 0 && hasLeadingZero[0] {
		return Str(DigitSeed, length)
	}

	if length == 1 {
		return Str(DigitWithoutZeroSeed, 1)
	}

	return Str(DigitWithoutZeroSeed, 1) + Str(DigitSeed, length-1)
}

// Symbols 生成指定长度的特殊字符串
func Symbols(length int) string {
	return Str(SymbolSeed, length)
}

// Lucky 根据概率抽取幸运值
func Lucky(probability float64, base ...float64) bool {
	if probability <= 0 {
		return false
	}

	b := float64(100)
	if len(base) > 0 {
		b = base[0]
	}

	if probability >= b {
		return true
	}

	str := strconv.FormatFloat(probability, 'f', -1, 64)
	scale := float64(0)

	if i := strings.IndexByte(str, '.'); i > 0 {
		scale = math.Pow10(len(str) - i - 1)
	}

	return Int64(1, int64(b*scale)) <= int64(probability*scale)
}

// Weight 权重随机
func Weight(fn func(v any) float64, list ...any) int {
	if len(list) == 0 {
		return -1
	}

	total := float64(0)
	scale := float64(1)

	for _, item := range list {
		weight := fn(item)
		str := strconv.FormatFloat(weight, 'f', -1, 64)

		if i := strings.IndexByte(str, '.'); i > 0 {
			scale = math.Max(scale, math.Pow10(len(str)-i-1))
		}

		total += weight
	}

	sum := int64(total * scale)

	if sum == 0 {
		return Int(1, len(list))
	}

	weight := Int64(1, sum)
	acc := int64(0)

	for i, item := range list {
		acc += int64(fn(item) * scale)
		if weight <= acc {
			return i
		}
	}

	return Int(1, len(list))
}

// Shuffle 打乱数组
func Shuffle(list []any) {
	globalRand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
}

func Rand() *rand.Rand {
	return globalRand
}
