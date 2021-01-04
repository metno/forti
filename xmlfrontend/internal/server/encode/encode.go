package encode

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"gitlab.met.no/forti/f2/internalprotocol"
	"gitlab.met.no/forti/f2/xmlfrontend/internal/server/config"
	"gitlab.met.no/forti/f2/xmlfrontend/pkg/xmlformat"
)

func Encode(location *internalprotocol.Location, forecast *internalprotocol.Forecast) (*xmlformat.ForecastDocument, error) {
	doc := xmlformat.ForecastDocument{
		XMLName: xml.Name{Local: "weatherdata"},
		Product: &xmlformat.ProductElement{
			Class: "pointData",
		},
		XSI:     "http://www.w3.org/2001/XMLSchema-instance",
		NsLoc:   "https://schema.api.met.no/schemas/weatherapi-0.4.xsd",
		Created: time.Now().UTC().Round(time.Second),
	}

	runEnded := forecast.Meta.UpdatedAt.AsTime().UTC().Round(time.Second)
	nextRun := forecast.Meta.NextUpdate.AsTime().UTC().Round(time.Second)

	sorted := sortByTime(forecast)
	if len(sorted) == 0 {
		return nil, errors.New("empty forecast")
	}

	var undef time.Time
	if sorted[0].Time == undef {
		altitude, ok := sorted[0].Values["altitude"]
		if ok {
			location.Altitude = &internalprotocol.Altitude{
				Value: altitude,
			}
		}
	}

	doc.Product = getProductElement(location, sorted)

	doc.Meta = &xmlformat.MetaElement{
		Models: []xmlformat.ModelElement{
			{
				Name:     "met_public_forecast",
				Termin:   runEnded.Round(time.Hour),
				Runended: runEnded,
				Nextrun:  nextRun,
				From:     doc.Product.Time[0].To,
				To:       doc.Product.Time[len(doc.Product.Time)-1].To,
			},
		},
	}

	return &doc, nil
}

type parsedForecast struct {
	Time   time.Time
	Values map[string]float32
}

func sortByTime(forecast *internalprotocol.Forecast) []parsedForecast {
	sorted := make(map[time.Time]map[string]float32)

	for _, data := range forecast.Data {
		for _, meta := range data.ParameterMeta {
			for i, t := range meta.Times {
				t := t.AsTime().UTC()
				values, ok := sorted[t]
				if !ok {
					values = make(map[string]float32)
					sorted[t] = values
				}
				values[meta.Parameter] = data.Data[int(meta.SliceFrom)+i]
			}
		}
	}

	ret := make([]parsedForecast, 0, len(sorted))
	for t, data := range sorted {
		f := parsedForecast{
			Time:   t,
			Values: data,
		}
		ret = append(ret, f)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Time.Before(ret[j].Time)
	})

	return ret
}

func getProductElement(location *internalprotocol.Location, forecast []parsedForecast) *xmlformat.ProductElement {
	return &xmlformat.ProductElement{
		Class: "pointData",
		Time:  getTimeElements(location, forecast),
	}
}

func getTimeElements(location *internalprotocol.Location, forecast []parsedForecast) []xmlformat.TimeElement {
	ret := make([]xmlformat.TimeElement, 0, len(forecast))

	for _, f := range forecast {
		timeElement := getElementsForTimestep(location, &f)
		ret = append(ret, timeElement...)
	}

	return ret
}

func getElementsForTimestep(location *internalprotocol.Location, forecast *parsedForecast) []xmlformat.TimeElement {

	var ret []xmlformat.TimeElement

	for _, elements := range config.Configuration.Elements {
		loc := getLocationElement(location, forecast, &elements)

		offset, err := time.ParseDuration(elements.Offset)
		if err != nil {
			panic(err)
		}

		fromTime := forecast.Time.Add(-offset)

		if config.Configuration.CutForecast {
			now := time.Now().Truncate(time.Hour)
			if forecast.Time.Before(now) {
				return nil
			}
		}

		if len(loc.Forecast) != 0 {
			f := xmlformat.TimeElement{
				Location: loc,
				DataType: "forecast",
				From:     fromTime,
				To:       forecast.Time,
			}
			ret = append(ret, f)
		}
	}
	return ret
}

func getLocationElement(location *internalprotocol.Location, forecast *parsedForecast, elements *config.DataElement) xmlformat.LocationElement {
	var altitude int
	if location.Altitude != nil {
		altitude = int(math.Round(float64(location.Altitude.Value)))
	}
	return xmlformat.LocationElement{
		Altitude:  altitude,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Forecast:  getDataElements(forecast, elements),
	}

}

func getDataElements(forecast *parsedForecast, elements *config.DataElement) []xmlformat.DataElement {
	var out []xmlformat.DataElement
	for _, p := range elements.Parameters {
		value, ok := forecast.Values[p.NcName]
		if !ok {
			continue
		}

		dataElement := xmlformat.DataElement{
			XMLName: xml.Name{
				Local: p.Name,
			},
		}
		for _, attr := range p.Attrs {
			dataElement.Attr = append(dataElement.Attr, xml.Attr{
				Name:  xml.Name{Local: attr.Name},
				Value: attr.Value,
			})
		}

		dataElement.Attr = append(dataElement.Attr, xml.Attr{
			Name:  xml.Name{Local: p.ValueName},
			Value: fmt.Sprintf("%0.1f", value),
		})

		for _, computed := range p.ComputedAttrs {
			f, ok := functions[computed.Func]
			if !ok {
				log.Println("missing function ", computed.Func)
				continue
			}
			value, err := f(value, &forecast.Values)
			if err != nil {
				log.Println(err)
				continue
			}
			dataElement.Attr = append(dataElement.Attr, xml.Attr{
				Name:  xml.Name{Local: computed.Name},
				Value: value,
			})

		}

		out = append(out, dataElement)
	}
	return out
}
