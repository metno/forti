package xmlformat

import (
	"encoding/xml"
	"time"
)

type ForecastDocument struct {
	XMLName xml.Name
	Meta    *MetaElement    `xml:"meta"`
	Product *ProductElement `xml:"product"`
	XSI     string          `xml:"xmlns:xsi,attr"`
	NsLoc   string          `xml:"xsi:noNamespaceSchemaLocation,attr"`
	Created time.Time       `xml:"created,attr"`
}

type MetaElement struct {
	Models []ModelElement `xml:"model"`
}

type ModelElement struct {
	Name     string    `xml:"name,attr"`
	Termin   time.Time `xml:"termin,attr"`
	Runended time.Time `xml:"runended,attr"`
	Nextrun  time.Time `xml:"nextrun,attr"`
	From     time.Time `xml:"from,attr"`
	To       time.Time `xml:"to,attr"`
}

type ProductElement struct {
	Time  []TimeElement `xml:"time"`
	Class string        `xml:"class,attr"`
}

type TimeElement struct {
	Location LocationElement `xml:"location"`
	DataType string          `xml:"datatype,attr"`
	From     time.Time       `xml:"from,attr"`
	To       time.Time       `xml:"to,attr"`
}

type LocationElement struct {
	Forecast  []DataElement `xml:",any"`
	Altitude  int           `xml:"altitude,attr"`
	Latitude  float32       `xml:"latitude,attr"`
	Longitude float32       `xml:"longitude,attr"`
}

type DataElement struct {
	XMLName xml.Name
	Attr    []xml.Attr `xml:",any,attr"`
}
