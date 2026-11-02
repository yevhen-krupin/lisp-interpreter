package core

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
