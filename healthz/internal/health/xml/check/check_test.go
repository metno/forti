package check

import (
	"encoding/xml"
	"fmt"
	"testing"

	"gitlab.met.no/forti/f2/healthz/internal/health/xml/config"
	"gitlab.met.no/forti/f2/xmlfrontend/pkg/xmlformat"
)

func TestXMLCheck(t *testing.T) {
	forecast := &xmlformat.ForecastDocument{}
	err := xml.Unmarshal([]byte(faultyXML), forecast)
	if err != nil {
		t.Errorf("decoding xml failed: %s", err)
	}

	var count uint = 3
	temperatureMin := float32(-90.0)
	temperatureMax := float32(90.0)

	blueprint := config.Blueprint{
		MaxAge: "1h",
		Parameters: map[string]config.CheckSpecification{
			"temperature":         {Attribute: "value", MinimumCount: &count, MinimumValue: &temperatureMin, MaximumValue: &temperatureMax},
			"windSpeed":           {Attribute: "mps", MinimumCount: &count},
			"windDirection":       {Attribute: "deg", MinimumCount: &count},
			"pressure":            {Attribute: "value", MinimumCount: &count},
			"humidity":            {Attribute: "value", MinimumCount: &count},
			"cloudiness":          {Attribute: "percent", MinimumCount: &count},
			"lowClouds":           {Attribute: "percent", MinimumCount: &count},
			"mediumClouds":        {Attribute: "percent", MinimumCount: &count},
			"highClouds":          {Attribute: "percent", MinimumCount: &count},
			"dewpointTemperature": {Attribute: "value", MinimumCount: &count},
			"windGust":            {Attribute: "mps", MinimumCount: &count},
			"areaMaxWindSpeed":    {Attribute: "mps", MinimumCount: &count},
			"fog":                 {Attribute: "percent", MinimumCount: &count},
			"symbol":              {Attribute: "number", MinimumCount: &count, Durations: []int{1, 6}},
			"precipitation":       {Attribute: "value", MinimumCount: &count, Durations: []int{1}},
		},
	}
	result := runChecks(forecast, blueprint)

	if result.OK {
		t.Error("expected errors when checking forecast")
	}

	expectedErrors := 5
	if len(result.Problems) != expectedErrors {
		errString := fmt.Sprintf("Expected %d errors in XML, got: %d. These are the errors: \n", expectedErrors, len(result.Problems))
		for _, e := range result.Problems {
			errString += fmt.Sprintln(e)
		}
		t.Errorf(errString)
	}
}

// Faulty because: Missing many, many timesteps, too old forecast and one wrong temperature value
const faultyXML = `<weatherdata xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://api.met.no/weatherapi/locationforecast/1.9/schema" created="2019-06-04T11:59:13Z">
   <meta>
      <model name="LOCAL" termin="2019-06-04T06:00:00Z" runended="2019-06-04T08:36:12Z" nextrun="2019-06-04T16:00:00Z" from="2019-06-04T12:00:00Z" to="2019-06-07T00:00:00Z" />
      <model name="EPS" termin="2019-06-04T00:00:00Z" runended="2019-06-04T09:09:11Z" nextrun="2019-06-04T22:00:00Z" from="2019-06-07T06:00:00Z" to="2019-06-13T18:00:00Z" />
   </meta>
   <product class="pointData">
      <time datatype="forecast" from="2019-06-04T21:00:00Z" to="2019-06-04T21:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
            <temperature id="TTT" unit="celsius" value="-273"/>
            <windDirection id="dd" deg="186.4" name="S"/>
            <windSpeed id="ff" mps="2.0" beaufort="2" name="Svak vind"/>
            <windGust id="ff_gust" mps="4.4"/>
            <areaMaxWindSpeed mps="2.8"/>
            <humidity value="75.1" unit="percent"/>
            <pressure id="pr" unit="hPa" value="1013.0"/>
            <cloudiness id="NN" percent="5.6"/>
            <fog id="FOG" percent="0.0"/>
            <lowClouds id="LOW" percent="0.0"/>
            <mediumClouds id="MEDIUM" percent="0.0"/>
            <highClouds id="HIGH" percent="5.6"/>
            <dewpointTemperature id="TD" unit="celsius" value="8.5"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T20:00:00Z" to="2019-06-04T21:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <symbol id="Sun" number="1"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T19:00:00Z" to="2019-06-04T21:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <symbol id="Sun" number="1"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T18:00:00Z" to="2019-06-04T21:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <symbol id="Sun" number="1"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T15:00:00Z" to="2019-06-04T21:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <minTemperature id="TTT" unit="celsius" value="12.8"/>
      <maxTemperature id="TTT" unit="celsius" value="18.4"/>
      <symbol id="LightCloud" number="2"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T22:00:00Z" to="2019-06-04T22:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
            <temperature id="TTT" unit="celsius" value="11.4"/>
            <windDirection id="dd" deg="185.8" name="S"/>
            <windSpeed id="ff" mps="1.7" beaufort="2" name="Svak vind"/>
            <windGust id="ff_gust" mps="3.6"/>
            <areaMaxWindSpeed mps="2.2"/>
            <humidity value="81.7" unit="percent"/>
            <pressure id="pr" unit="hPa" value="1013.3"/>
            <cloudiness id="NN" percent="28.7"/>
            <fog id="FOG" percent="0.0"/>
            <lowClouds id="LOW" percent="0.0"/>
            <mediumClouds id="MEDIUM" percent="0.0"/>
            <highClouds id="HIGH" percent="28.7"/>
            <dewpointTemperature id="TD" unit="celsius" value="8.4"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T21:00:00Z" to="2019-06-04T22:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <symbol id="LightCloud" number="2"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T20:00:00Z" to="2019-06-04T22:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <symbol id="LightCloud" number="2"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T19:00:00Z" to="2019-06-04T22:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
            <precipitation unit="mm" value="0.0" minvalue="0.0" maxvalue="0.0"/>
      <symbol id="Sun" number="1"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T16:00:00Z" to="2019-06-04T22:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <minTemperature id="TTT" unit="celsius" value="11.4"/>
      <maxTemperature id="TTT" unit="celsius" value="18.4"/>
      <symbol id="LightCloud" number="2"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T23:00:00Z" to="2019-06-04T23:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
            <temperature id="TTT" unit="celsius" value="10.3"/>
            <windDirection id="dd" deg="183.0" name="S"/>
            <windSpeed id="ff" mps="1.1" beaufort="1" name="Flau vind"/>
            <windGust id="ff_gust" mps="3.0"/>
            <areaMaxWindSpeed mps="2.0"/>
            <humidity value="87.3" unit="percent"/>
            <pressure id="pr" unit="hPa" value="1013.6"/>
            <cloudiness id="NN" percent="30.8"/>
            <lowClouds id="LOW" percent="0.0"/>
            <mediumClouds id="MEDIUM" percent="0.0"/>
            <highClouds id="HIGH" percent="30.7"/>
            <dewpointTemperature id="TD" unit="celsius" value="8.3"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T22:00:00Z" to="2019-06-04T23:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <symbol id="LightCloud" number="2"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T21:00:00Z" to="2019-06-04T23:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <symbol id="LightCloud" number="2"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T20:00:00Z" to="2019-06-04T23:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <symbol id="LightCloud" number="2"/>
         </location>
      </time>
      <time datatype="forecast" from="2019-06-04T17:00:00Z" to="2019-06-04T23:00:00Z">
         <location altitude="189" latitude="60.1427" longitude="11.1630">
      <minTemperature id="TTT" unit="celsius" value="10.3"/>
      <maxTemperature id="TTT" unit="celsius" value="18.1"/>
      <symbol id="Sun" number="1"/>
         </location>
      </time>
	</product>
</weatherdata>
`
