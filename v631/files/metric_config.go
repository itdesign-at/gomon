package files

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/itdesign-at/golib/keyvalue"
)

const f_metricConfigFile = "/opt/watchit/var/etc/metricConfig.json"

type MetricConfig struct {
	baseConfig keyvalue.Record
	data       map[string]keyvalue.Record
}

func NewMetricConfig() *MetricConfig {
	var mc MetricConfig
	bc := newBaseConfig()
	mc.baseConfig = bc.getAll()
	return &mc
}

func (mc *MetricConfig) Get(key string) (keyvalue.Record, error) {
	var ret, tmp keyvalue.Record

	if mc.data == nil {
		content, err := os.ReadFile(f_metricConfigFile)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(content, &mc.data)
		if err != nil {
			panic(err)
		}
	}

	var ok bool
	if tmp, ok = mc.data[key]; !ok {
		return nil, fmt.Errorf("%s not found", key)
	}

	// avoid pointer manipulation by copying the record
	ret = tmp.Copy()

	re := regexp.MustCompile(`{{\.(.*?)}}`)
	for k, v := range ret {
		switch value := v.(type) {
		case string:
			if !strings.Contains(value, "{{.") {
				continue
			}
			myMatches := re.FindAllStringSubmatch(value, -1)
			for _, m := range myMatches {
				from := m[0]
				baseKey := strings.Replace(m[1], "BC.", "", 1)
				to := mc.baseConfig.String(baseKey)
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

// GetFromBase returns a string value from the base config file
func (mc *MetricConfig) GetFromBase(key string) (string, error) {
	if mc.baseConfig == nil {
		return "", fmt.Errorf("base config not initialized")
	}
	if !mc.baseConfig.Exists(key) {
		return "", fmt.Errorf("%s not found", key)
	}
	return mc.baseConfig.String(key, true), nil
}
