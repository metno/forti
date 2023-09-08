package modelprovider

import (
	"fmt"
)

// ID contains the parsed data of an state storage's data entry key
type ID struct {
	Param   string `json:"parameter"`
	Group   string `json:"group"`
	Version int    `json:"version"`
}

func (i ID) String() string {
	return fmt.Sprintf("%v/%v/%v", i.Param, i.Group, i.Version)
}
