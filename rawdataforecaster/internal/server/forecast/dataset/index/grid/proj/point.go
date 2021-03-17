package proj

import "fmt"

// Point specifies a position in a 2D space
type Point struct {
	X float64
	Y float64
}

// WKT returns the Well-Known Text representation of this coordinate
func (p Point) WKT() string {
	return fmt.Sprintf("POINT(%f %f)", p.X, p.Y)
}

func (p Point) String() string {
	return p.WKT()
}
