package files

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/itdesign-at/golib/keyvalue"
)

const f_hostsExportedFile = "/opt/watchit/var/etc/hosts-exported.json"

var (
	ErrHostNotFound = errors.New("host not found")
)

type HostsExported struct {
	data map[string]keyvalue.Record
}

func NewHostsExported() *HostsExported {
	var he *HostsExported
	return he
}

func (he *HostsExported) Read() error {
	content, err := os.ReadFile(f_hostsExportedFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &he.data)
	if err != nil {
		return err
	}
	return nil
}

func (he *HostsExported) GetHostProperties(host string) (keyvalue.Record, error) {
	var err error
	var ret keyvalue.Record
	var ok bool
	if he.data == nil {
		err = he.Read()
		if err != nil {
			return nil, err
		}
	}
	if ret, ok = he.data[host]; !ok {
		return nil, ErrHostNotFound
	}
	return ret, nil
}
