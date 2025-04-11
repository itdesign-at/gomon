package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/itdesign-at/golib/keyvalue"
)

var baseConfig keyvalue.Record
var metricConfig map[string]keyvalue.Record

func GetMetricDefinition(wantedKey string) (keyvalue.Record, error) {

	var ret, tmp keyvalue.Record

	// avoid pointer manipulation by copying the record
	theBaseConfig := GetBase().Copy()

	if metricConfig == nil {
		content, err := os.ReadFile(F_metricConfigFile)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(content, &metricConfig)
		if err != nil {
			panic(err)
		}
	}

	var ok bool
	if tmp, ok = metricConfig[wantedKey]; !ok {
		return nil, fmt.Errorf("%s not found", wantedKey)
	}

	// avoid pointer manipulation by copying the record
	ret = tmp.Copy()

	re := regexp.MustCompile(`\{\{\.(.*?)\}\}`)
	for k, v := range ret {
		switch value := v.(type) {
		case string:
			if !strings.Contains(value, "{{.") {
				continue
			}
			myMatches := re.FindAllStringSubmatch(value, -1)
			for _, m := range myMatches {
				from := m[0]
				key := strings.Replace(m[1], "BC.", "", 1)
				to := theBaseConfig.String(key)
				value = strings.Replace(value, from, to, -1)
			}
			ret[k] = value
		default:
			// for future use - allow other than string
			// in the config file
		}
	}
	return ret, nil
}

func GetBase() keyvalue.Record {

	if baseConfig != nil {
		return baseConfig
	}

	content, err := os.ReadFile(F_baseConfigFile)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content, &baseConfig)
	if err != nil {
		panic(err)
	}
	// Pattern to match [[.Something]]
	re := regexp.MustCompile(`\[\[\.(\w+)\]\]`)
	for k, v := range baseConfig {
		switch value := v.(type) {
		case string:
			if !strings.Contains(value, "[[.") {
				continue
			}
			myMatches := re.FindAllStringSubmatch(value, -1)
			for _, m := range myMatches {
				from := m[0]
				to := baseConfig.String(m[1])
				value = strings.Replace(value, from, to, -1)
			}
			baseConfig[k] = value
		default:
			// for future use - allow other than string
			// in the config file
		}
	}
	return baseConfig
}
