package v631

import (
	"encoding/json"
	"os"

	"github.com/itdesign-at/golib/keyvalue"
)

type HostsExported struct {
	data keyvalue.Record
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
