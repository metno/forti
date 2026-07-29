package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/invopop/jsonschema"
	"github.com/metno/forti/jsonfrontend/pkg/jsonformat"
)

func main() {
	version := flag.String("version", "", "semantic version to embed in the schema $id (e.g. v1.0.0)")
	flag.Parse()

	schema := jsonschema.Reflect(&jsonformat.GeoJSON{})

	if *version != "" {
		schema.ID = jsonschema.ID(fmt.Sprintf(
			"https://raw.githubusercontent.com/metno/forti/forecast.schema.%s.json",
			*version,
		))
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}
