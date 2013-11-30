package Db

import "strconv"

func strToBool(str string) bool {
	b, _ := strconv.ParseBool(str)
	return b
}

func strToFloat32(str string) float32 {
	v, _ := strconv.ParseFloat(str, 32)
	return float32(v)
}

func strToFloat64(str string) float64 {
	v, _ := strconv.ParseFloat(str, 64)
	return v
}

func strToInt(str string) int {
	v, _ := strconv.ParseInt(str, 10, 32)
	return int(v)
}

func strToInt8(str string) int8 {
	v, _ := strconv.ParseInt(str, 10, 8)
	return int8(v)
}

func strToInt16(str string) int16 {
	v, _ := strconv.ParseInt(str, 10, 16)
	return int16(v)
}

func strToInt32(str string) int32 {
	v, _ := strconv.ParseInt(str, 10, 32)
	return int32(v)
}

func strToInt64(str string) int64 {
	v, _ := strconv.ParseInt(str, 10, 64)
	return int64(v)
}

func strToUint(str string) uint {
	v, _ := strconv.ParseUint(str, 10, 32)
	return uint(v)
}

func strToUint8(str string) uint8 {
	v, _ := strconv.ParseUint(str, 10, 8)
	return uint8(v)
}

func strToUint16(str string) uint16 {
	v, _ := strconv.ParseUint(str, 10, 16)
	return uint16(v)
}

func strToUint32(str string) uint32 {
	v, _ := strconv.ParseUint(str, 10, 32)
	return uint32(v)
}

func strToUint64(str string) uint64 {
	v, _ := strconv.ParseUint(str, 10, 64)
	return uint64(v)
}

func snakeCasedName(name string) string {
	newStr := make([]rune, 0)
	for idx, chr := range name {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if idx > 0 {
				newStr = append(newStr, '_')
			}
			chr -= ('A' - 'a')
		}
		newStr = append(newStr, chr)
	}
	return string(newStr)
}
