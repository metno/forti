package mox

import (
	"encoding/xml"
	"time"
)

type ForecastDocument struct {
	XMLName xml.Name   `xml:"mox:Forecasts"`
	Attr    []xml.Attr `xml:",any,attr"`

	Description      string     `xml:"gml:description,omitempty"`
	Procedure        *Reference `xml:"mox:procedure,omitempty"`
	ObservedProperty *Reference `xml:"mox:observedProperty,omitempty"`

	ForecastPoint *ForecastPoint `xml:"mox:forecastPoint,omitempty"`
	IssueTime     *GmlTime       `xml:"mox:issueTime,omitempty"`
	NextIssueTime *GmlTime       `xml:"mox:nextIssueTime,omitempty"`

	CollectedForecast []CollectedForecast `xml:"mox:forecast,omitempty"`
}

type Reference struct {
	HREF string `xml:"xlink:href,attr"`
}

type ForecastPoint struct {
	Point GmlPoint `xml:"gml:Point"`
}

type GmlPoint struct {
	GmlID   string `xml:"gml:id,attr"`
	SrsName string `xml:"srsName,attr"`
	Pos     string `xml:"gml:pos"`
}

type GmlTime struct {
	TimeInstant GmlTimeInstant `xml:"gml:TimeInstant"`
}

type GmlTimeInstant struct {
	GmlID        string    `xml:"gml:id,attr"`
	TimePosition time.Time `xml:"gml:timePosition"`
}

type CollectedForecast struct {
	OceanForecast *OceanForecast `xml:"metno:OceanForecast"`
}

type OceanForecast struct {
	GmlID     string        `xml:"gml:id,attr"`
	ValidTime GmlTimePeriod `xml:"mox:validTime>gml:TimePeriod"`

	SeaBottomTopography        *Data `xml:"mox:seaBottomTopography,omitempty"`
	SeaIcePresence             *Data `xml:"mox:seaIcePresence,omitempty"`
	MeanTotalWaveDirection     *Data `xml:"mox:meanTotalWaveDirection,omitempty"`
	SignificantTotalWaveHeight *Data `xml:"mox:significantTotalWaveHeight,omitempty"`
	SeaCurrentDirection        *Data `xml:"mox:seaCurrentDirection,omitempty"`
	SeaCurrentSpeed            *Data `xml:"mox:seaCurrentSpeed,omitempty"`
	SeaTemperature             *Data `xml:"mox:seaTemperature,omitempty"`
}

type Data struct {
	UOM string `xml:"uom,attr"`

	// Value is the numeric value of the data. It is always a number, but it is not possible to for innerxml to be a float.
	Value string `xml:",innerxml"`
}

type GmlTimePeriod struct {
	GmlID string    `xml:"gml:id,attr"`
	Begin time.Time `xml:"gml:begin"`
	End   time.Time `xml:"gml:end"`
}
