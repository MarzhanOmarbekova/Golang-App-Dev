package main

import (
	"fmt"
	"math"
)

func Sqrt(x float64) float64 {
	z := 1.0
	var z_last float64
	for {
		z_last = z
		z -= (z*z - x) / (2 * z)
		fmt.Println(z, "\n")

		if math.Abs(z_last-z) < 1e-10 {
			break
		}
	}
	return z
}

func main() {
	fmt.Println(Sqrt(10024))
	fmt.Println(math.Sqrt(10024))
}
