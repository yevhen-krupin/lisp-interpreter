package core

import (
	"slices"
)

func int32_to_bytes(i int) []byte {
	buf := make([]byte, 4)
	buf[3] = byte(i & 0xFF)
	buf[2] = byte((i & 0xFF00) >> 8)
	buf[1] = byte((i & 0xFF0000) >> 16)
	buf[0] = byte((i & 0xFF000000) >> 24)
	return buf
}

func bytes_to_int32(bytes []byte) int {
	v := int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
	return v
}

func looks_like_string(data []byte, markers []string) bool {
	// starts and ends with one of string markers
	return len(data) >= 2 && slices.Contains(markers, string(data[0])) && slices.Contains(markers, string(data[len(data)-1]))
}

func value_to_any(data []byte) any {
	if len(data) == 1 {
		if data[0] == 0 {
			return false
		}
		if data[0] == 1 {
			return true
		}
	}
	if len(data) != 4 {
		return string(data)
	}
	for _, ch := range data {
		// only bytes of real numbers, not characters meaning digits
		if ch > 9 {
			return string(data)
		}
	}
	return bytes_to_int32(data)
}

func coalesce[TIn any, TOut comparable](in TIn, nodes ...func(TIn) TOut) TOut {
	var zero TOut
	for _, f := range nodes {
		n := f(in)
		if n != zero {
			return n
		}
	}
	return zero
}
