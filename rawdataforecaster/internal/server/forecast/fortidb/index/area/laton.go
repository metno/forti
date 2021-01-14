package area

import "fmt"

// LatLon specifies a latitude/longitude pair
type LatLon struct {
	Longitude float64
	Latitude  float64
}

// WKT returns the Well-Known Text representation of this coordinate
func (ll LatLon) WKT() string {
	return fmt.Sprintf("POINT(%f %f)", ll.Longitude, ll.Latitude)
}

func (ll LatLon) String() string {
	return ll.WKT()
}
