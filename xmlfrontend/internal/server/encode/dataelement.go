package encode

import (
	"encoding/xml"
	"log"
	"strconv"

	"github.com/metno/forti/xmlfrontend/internal/server/config"
	"github.com/metno/forti/xmlfrontend/internal/xmlformat"
)

func newDataElement(data *config.Parameter, value float32, allValues *map[string]float32) (xmlformat.DataElement, error) {
	element := xmlformat.DataElement{XMLName: xml.Name{Local: data.Name}}

	for _, attr := range data.Attrs {
		attribute := xml.Attr{
			Name:  xml.Name{Local: attr.Name},
			Value: attr.Value,
		}
		element.Attr = append(element.Attr, attribute)
	}
	if data.ValueName != "" {
		valueAttr := xml.Attr{
			Name:  xml.Name{Local: data.ValueName},
			Value: strconv.FormatFloat(float64(value), 'f', 1, 32),
		}
		element.Attr = append(element.Attr, valueAttr)
	}
	for _, cAttr := range data.ComputedAttrs {
		fun, ok := functions[cAttr.Func]
		if !ok {
			log.Println("missing function " + cAttr.Func)
			continue
		}
		value, err := fun(value, allValues)
		if err != nil {
			//log.Println(err)
			continue
		}
		attribute := xml.Attr{
			Name:  xml.Name{Local: cAttr.Name},
			Value: value,
		}
		element.Attr = append(element.Attr, attribute)
	}

	return element, nil
}
