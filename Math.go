package zgo

import (
	"math"
)

//  Created by Tor Langballe on /23/9/14.

const MathPI float64 = 3.141592653589793

const MathDegreesToMeters = (111.32 * 1000)
const MathMetersToDegrees = 1 / MathDegreesToMeters

func MathRadToDeg(rad float64) float64 {
	return rad * 180 / MathPI
}

func MathDegToRad(deg float64) float64 {
	return deg * MathPI / 180
}

func MathAngleDegToPos(deg float64) Pos {
	return Pos{math.Sin(MathDegToRad(deg)), -math.Cos(MathDegToRad(deg))}
}
func MathPosToAngleDeg(pos Pos) float64 {
	return MathRadToDeg(MathArcTanXYToRad(pos))
}
func MathGetDistanceFromLongLatInMeters(pos1 Pos, pos2 Pos) float64 {
	R := 6371.0 // Radius of the earth in km
	dLat := MathDegToRad(pos2.Y - pos1.Y)
	dLon := MathDegToRad(pos2.X - pos1.X)
	a := (math.Pow(math.Sin(dLat/2.0), 2.0) + math.Cos(MathDegToRad(pos1.Y))) * math.Cos(MathDegToRad(pos2.Y)) * math.Pow(math.Sin(dLon/2.0), 2.0)
	c := 2.0 * float64(math.Asin(math.Sqrt(math.Abs(a))))
	return c * R * 1000.0
}

func MathFraction(v float64) float64 {
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

func MathArcTanXYToRad(pos Pos) float64 {
	var a = float64(math.Atan2(pos.Y, pos.X))
	if a < 0 {
		a += MathPI * 2
	}
	return a
}

func MathMixedArrayValueAtIndex(array []float64, index float64) float64 {
	if index < 0.0 {
		return array[0]
	}
	if index >= float64(len(array))-1 {
		return array[len(array)-1]
	}
	n := index
	f := (index - n)
	var v = array[int(n)] * (1 - f)
	if int(n) < len(array) {
		v += array[int(n+1)] * f
		return v
	}
	if len(array) > 0 {
		return array[len(array)-1]
	}
	return 0
}

func MathMixedArrayValueAtT(array []float64, t float64) float64 {
	return MathMixedArrayValueAtIndex(array, float64(len(array)-1)*t)
}
