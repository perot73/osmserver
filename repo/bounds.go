package repo

import (
	"strconv"
	"strings"
)

type Bounds struct {
	Minlon float64 `xml:"minlon,attr"`
	Minlat float64 `xml:"minlat,attr"`
	Maxlon float64 `xml:"maxlon,attr"`
	Maxlat float64 `xml:"maxlat,attr"`
}

func NewBounds(bbox string) (*Bounds, error) {
	bounds := &Bounds{0, 0, 0, 0}
	var err error
	coords := strings.Split(bbox, ",")

	bounds.Minlon, err = strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return bounds, err
	}

	bounds.Minlat, err = strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return bounds, err
	}

	bounds.Maxlon, err = strconv.ParseFloat(coords[2], 64)
	if err != nil {
		return bounds, err
	}

	bounds.Maxlat, err = strconv.ParseFloat(coords[3], 64)
	if err != nil {
		return bounds, err
	}

	return bounds, nil
}
