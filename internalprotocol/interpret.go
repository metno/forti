package internalprotocol

import "time"

type InterpretedData struct {
	Times  []time.Time
	Values []float32
}

func InterpretValues(forecast *Forecast) map[string]InterpretedData {
	ret := make(map[string]InterpretedData)

	for _, meta := range forecast.ParameterMeta {
		times := make([]time.Time, len(meta.Times))
		for i, t := range meta.Times {
			times[i] = t.AsTime().UTC()
		}
		d := InterpretedData{
			Times:  times,
			Values: forecast.Data[meta.SliceFrom : int(meta.SliceFrom)+len(times)],
		}
		ret[meta.Parameter] = d
		if len(d.Times) != len(d.Values) {
			panic("size mismatch")
		}
	}
	return ret
}

func SortByTime(interpretedData map[string]InterpretedData, parameters ...string) map[time.Time]map[string]float32 {
	ret := make(map[time.Time]map[string]float32)

	for _, parameter := range parameters {
		id, ok := interpretedData[parameter]
		if !ok {
			continue
		}
		for i, t := range id.Times {
			values, ok := ret[t]
			if !ok {
				values = make(map[string]float32)
				ret[t] = values
			}
			values[parameter] = id.Values[i]
		}
	}

	return ret
}
