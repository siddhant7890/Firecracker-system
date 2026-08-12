// Package geo provides simple geographic distance calculations.
package geo

import (
	"fmt"
	"math"
)

const earthRadiusMeters = 6371000

// DistanceMeters returns the great-circle distance between two lat/lng
// points, in meters, using the haversine formula.
func DistanceMeters(lat1, lng1, lat2, lng2 float64) float64 {
	rad := func(deg float64) float64 { return deg * math.Pi / 180 }

	dLat := rad(lat2 - lat1)
	dLng := rad(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}

// CheckWithinRadius returns an error if (lat, lng) is farther than
// radiusMeters from (centerLat, centerLng).
func CheckWithinRadius(lat, lng, centerLat, centerLng, radiusMeters float64) error {
	dist := DistanceMeters(lat, lng, centerLat, centerLng)
	if dist > radiusMeters {
		return fmt.Errorf("location is %.0fm away, must be within %.0fm", dist, radiusMeters)
	}
	return nil
}
