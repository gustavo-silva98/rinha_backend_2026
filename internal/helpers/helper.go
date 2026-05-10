package helpers

import "rinha2026/internal/preprocess"

func DistEuclid(a []int8, b []int8) int32 {
	var acc int32
	for i := 0; i < preprocess.Dims; i++ {
		d := int32(a[i]) - int32(b[i])
		acc += d * d
	}
	return acc
}
