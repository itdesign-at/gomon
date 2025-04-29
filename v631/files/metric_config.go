package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/itdesign-at/golib/converter"
	"github.com/itdesign-at/golib/keyvalue"
)

const f_metricConfigFile = "/opt/watchit/var/etc/metricConfig.json"

type MetricConfig struct {
	baseConfig keyvalue.Record
	data       map[string]keyvalue.Record
	macros     keyvalue.Record
}

// NewMetricConfig creates a new MetricConfig instance and
// initializes the baseConfig from the baseConfig.json file.
func NewMetricConfig() *MetricConfig {
	var mc MetricConfig
	mc.macros = keyvalue.NewRecord()
	bc := newBaseConfig()
	mc.baseConfig = bc.getAll()
	return &mc
}

// WithMacros add text template macros in normalized form
func (mc *MetricConfig) WithMacros(macros keyvalue.Record) *MetricConfig {

	var get = func(key string) string {
		switch key {
		case "k":
			return macros.String("k")
		case "n":
			if macros.Exists("n") {
				return converter.Normalize(macros.String("n"))
			} else {
				return converter.Normalize(macros.String("Node"))
			}
		case "h":
			if macros.Exists("h") {
				return converter.Normalize(macros.String("h"))
			} else {
				return converter.Normalize(macros.String("Host"))
			}
		case "s":
			if macros.Exists("s") {
				return converter.Normalize(macros.String("s"))
			} else {
				return converter.Normalize(macros.String("Service"))
			}
		}
		return ""
	}

	mc.AddMacro("K", get("k"))
	mc.AddMacro("H", get("h"))
	mc.AddMacro("S", get("s"))

	nhs := []string{
		get("n"),
		mc.macros.String("H"),
		mc.macros.String("S"),
	}
	mc.AddMacro("NHS", strings.Join(nhs, "/"))
	return mc
}

// AddMacro adds a single text template macro
func (mc *MetricConfig) AddMacro(key string, value any) *MetricConfig {
	mc.macros[key] = value
	return mc
}

// Get returns one metric configuration for the given key
// like e.g. "binary". It replaces text template macros when
// the macro map is not empty.
func (mc *MetricConfig) Get(key string) (keyvalue.Record, error) {
	var ret, tmp keyvalue.Record

	if mc.data == nil {
		content, err := os.ReadFile(f_metricConfigFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(content, &mc.data)
		if err != nil {
			return nil, err
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
			// replace .BC macros first
			myMatches := re.FindAllStringSubmatch(value, -1)
			for _, m := range myMatches {
				from := m[0]
				baseKey := strings.Replace(m[1], "BC.", "", 1)
				to := mc.baseConfig.String(baseKey)
				value = strings.Replace(value, from, to, -1)
			}

			// replace other macros like e.g. {{.NHS}} which
			// are stored in the mc.macros map
			if strings.Contains(value, "{{.") && len(mc.macros) > 0 {
				var b bytes.Buffer
				tpl, _ := template.New("x").Parse(value)
				tpl.Option("missingkey=zero")
				_ = tpl.Execute(&b, mc.macros)
				value = b.String()
			}
			ret[k] = value
		default:
			// for future use - allow other than string
			// in the config file
		}
	}

	// add $nats_subject
	if nhs := mc.macros.String("NHS"); nhs != "" {
		var versionField = ""
		// for future use - check V1, V2, V3, etc.
		for i := 1; i <= 10; i++ {
			str := "V" + strconv.Itoa(i)
			if strings.Contains(ret.String("$from"), "/"+str+"/") {
				versionField = str
				break
			}
		}
		ret["$nats_subject"] = versionField + "." +
			ret.String("$metric") + "." +
			strings.Replace(nhs, "/", ".", -1)
	}

	return ret, nil
}

// GetFromBase returns a string value from the baseConfig.json file
func (mc *MetricConfig) GetFromBase(key string) (string, error) {
	if mc.baseConfig == nil {
		return "", fmt.Errorf("base config not initialized")
	}
	if !mc.baseConfig.Exists(key) {
		return "", fmt.Errorf("%s not found", key)
	}
	return mc.baseConfig.String(key, true), nil
}
