package encode

import (
	"encoding/xml"
	"fmt"
	"sort"
	"time"

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/moxfrontend/pkg/mox"
)

func EncodeNoData() *mox.ForecastDocument {
	return &mox.ForecastDocument{
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "gml:id"}, Value: time.Now().UTC().Format("pt0-20060102")},
			{Name: xml.Name{Local: "xmlns:mox"}, Value: "http://wdb.met.no/wdbxml"},
			{Name: xml.Name{Local: "xmlns:metno"}, Value: "http://api.met.no"},
			{Name: xml.Name{Local: "xmlns:gml"}, Value: "http://www.opengis.net/gml"},
			{Name: xml.Name{Local: "xmlns:xlink"}, Value: "http://www.w3.org/1999/xlink"},
			{Name: xml.Name{Local: "xmlns:xsi"}, Value: "http://www.w3.org/2001/XMLSchema-instance"},
			{Name: xml.Name{Local: "xsi:schemaLocation"}, Value: "http://wdb.met.no/wdbxml/schema/products.xsd"},
		},
	}
}

func Encode(forecast *internalprotocol.Forecast) (*mox.ForecastDocument, error) {

	location := forecast.ForecastMeta.GridLocation

	ret := emptyDocument(
		location.Longitude, location.Latitude, -2147483648,
		forecast.ForecastMeta.UpdatedAt.AsTime(),
		forecast.ForecastMeta.NextUpdate.AsTime(),
	)

	var depth string = "0.0"
	for _, p := range forecast.ParameterMeta {
		if p.Parameter == "sea_floor_depth_below_sea_level" {
			depth = fmt.Sprintf("%.1f", forecast.Data[p.SliceFrom])
		}
	}

	// cutoff time for forecast
	lastTime := forecast.ForecastMeta.UpdatedAt.AsTime().Add(5 * 24 * time.Hour)
	lastTime = time.Date(lastTime.Year(), lastTime.Month(), lastTime.Day(), 0, 0, 0, 0, time.UTC)

	steps := make(map[time.Time]*mox.OceanForecast)
	for _, meta := range forecast.ParameterMeta {
		for timeIndex, t := range meta.Times {
			if t.AsTime().After(lastTime) {
				break
			}

			step, ok := steps[t.AsTime()]
			if !ok {
				step = &mox.OceanForecast{
					ValidTime: mox.GmlTimePeriod{
						Begin: t.AsTime().UTC().Round(time.Second),
						End:   t.AsTime().UTC().Round(time.Second),
					},
					SeaBottomTopography: &mox.Data{
						UOM:   "m",
						Value: depth,
					},
				}
				steps[t.AsTime()] = step
			}
			value := forecast.Data[int(meta.SliceFrom)+timeIndex]
			format := func(value float32) string {
				return fmt.Sprintf("%.1f", value)
			}
			switch meta.Parameter {
			case "sea_ice_area_fraction":
				step.SeaIcePresence = &mox.Data{
					UOM:   "%",
					Value: fmt.Sprintf("%d", int(forecast.Data[int(meta.SliceFrom)+timeIndex])),
				}
			case "sea_surface_wave_from_direction":
				step.MeanTotalWaveDirection = &mox.Data{
					UOM:   "deg",
					Value: format(flipAngle(value)),
				}
			case "sea_surface_wave_height":
				step.SignificantTotalWaveHeight = &mox.Data{
					UOM:   "m",
					Value: format(value),
				}
			case "sea_water_speed":
				step.SeaCurrentSpeed = &mox.Data{
					UOM:   "m/s",
					Value: format(value),
				}
			case "sea_water_temperature":
				step.SeaTemperature = &mox.Data{
					UOM:   "Cel",
					Value: format(value),
				}
			case "sea_water_to_direction":
				step.SeaCurrentDirection = &mox.Data{
					UOM:   "deg",
					Value: format(value),
				}
			}
		}
	}

	var elements []mox.CollectedForecast
	for _, step := range steps {
		elements = append(elements, mox.CollectedForecast{step})
	}
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].OceanForecast.ValidTime.Begin.Before(elements[j].OceanForecast.ValidTime.Begin)
	})

	// now := time.Now().Truncate(time.Hour).Add(time.Hour)
	now := time.Now()

	for i, e := range elements {
		if e.OceanForecast.ValidTime.Begin.After(now) {
			if i > 0 {
				elements = elements[i:]
			}
			break
		}
	}

	for i, e := range elements {
		e.OceanForecast.ValidTime.GmlID = fmt.Sprintf("vt-%d", i)
		e.OceanForecast.GmlID = fmt.Sprintf("f-%d", i)
	}

	ret.CollectedForecast = elements

	return ret, nil
}

func flipAngle(angle float32) float32 {
	angle += 180
	if angle > 360 {
		angle -= 360
	}
	return angle
}

func emptyDocument(longitude, latitude float32, depth int, issueTime time.Time, nextIssueTime time.Time) *mox.ForecastDocument {
	doc := EncodeNoData()
	doc.Description = "Location forecast from api.met.no "
	doc.Procedure = &mox.Reference{HREF: "http://api.met.no/yr-procedure-desc.html"}
	doc.ObservedProperty = &mox.Reference{HREF: "urn:x-ogc:def:phenomenon:weather"}
	doc.ForecastPoint = &mox.ForecastPoint{
		Point: mox.GmlPoint{
			GmlID:   "fp-0",
			SrsName: "urn:ogc:def:crs:epsg:4326",
			Pos:     fmt.Sprintf("%.4f %.4f %d", longitude, latitude, depth),
		},
	}
	doc.IssueTime = &mox.GmlTime{
		TimeInstant: mox.GmlTimeInstant{
			GmlID:        "it-0",
			TimePosition: issueTime.UTC().Round(time.Second),
		},
	}
	doc.NextIssueTime = &mox.GmlTime{
		TimeInstant: mox.GmlTimeInstant{
			GmlID:        "nit-0",
			TimePosition: nextIssueTime.UTC().Add(time.Hour).Truncate(time.Hour),
		},
	}

	return doc
}
