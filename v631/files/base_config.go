package files

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/itdesign-at/golib/keyvalue"
)

const f_baseConfigFile = "/opt/watchit/var/etc/baseConfig.json"

type baseConfig struct {
	data keyvalue.Record
}

// newBaseConfig is required to read the baseConfig.json file
func newBaseConfig() *baseConfig {
	var bc baseConfig
	return &bc
}

// read reads the baseConfig.json file and replaces all
// occurrences of [[.Something]] with the value of the
// corresponding key in the baseConfig.json file.
func (bc *baseConfig) read() error {

	if bc.data != nil { // already read?
		return nil
	}

	content, err := os.ReadFile(f_baseConfigFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &bc.data)
	if err != nil {
		return err
	}
	// Pattern to match [[.Something]]
	re := regexp.MustCompile(`\[\[\.(\w+)]]`)
	for k, v := range bc.data {
		switch value := v.(type) {
		case string:
			if !strings.Contains(value, "[[.") {
				continue
			}
			myMatches := re.FindAllStringSubmatch(value, -1)
			for _, m := range myMatches {
				from := m[0]
				to := bc.data.String(m[1])
				value = strings.Replace(value, from, to, -1)
			}
			bc.data[k] = value
		default:
			// for future use - allow other than string
			// in the config file
		}
	}
	return nil
}

// getAll returns all key/value pairs as a copy
// of the baseConfig.json file
func (bc *baseConfig) getAll() keyvalue.Record {
	_ = bc.read()
	return bc.data.Copy()
}

// getString returns the string value for a specific key
func (bc *baseConfig) getString(key string) (string, error) {
	if bc.data == nil {
		err := bc.read()
		if err != nil {
			return "", err
		}
	}
	if !bc.data.Exists(key) {
		return "", errors.New("key not found")
	}
	return bc.data.String(key, true), nil
}
