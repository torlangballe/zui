package zgo

import (
	"math"
)

//  Created by Tor Langballe on /23/9/14.

// const MathPI float64 = 3.141592653589793


/*
func MathFraction(v float64) float64 { // use math.Modf
	return v - MathFloor(v)
}

func MathFloor(v float64) float64 {
	return math.Floor(v)
}

func MathCeil(v float64) float64 {
	return math.Ceil(v)
}

func MathLog10(d float64) float64 {
	return math.Log10(d)
}
*/

func MathGetNiceIncsOf(d float64, incCount int, isMemory bool) float64 {
	l := math.Floor(math.Log10(d))
	var n = math.Pow(10.0, l)
	if isMemory {
		n = math.Pow(1024.0, math.Ceil(l/3.0))
		for d/n > float64(incCount) {
			n = n * 2.0
		}
	}
	for d/n < float64(incCount) {
		n = n / 2.0
	}
	return n
}


// func MathBitwiseInvert(v uint64) uint64 {
// 	return ^v
// }
