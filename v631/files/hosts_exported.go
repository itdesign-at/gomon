package files

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/gosnmp/gosnmp"
)

const f_hostsExportedFile = "/opt/watchit/var/etc/hosts-exported.json"

var (
	ErrHostNotFound = errors.New("host not found")
)

type HostProperties struct {
	Host         string `json:"h"`
	Description  string `json:"D"`
	IPAddress    string `json:"IP"`
	Community    string `json:"c"`
	Version      string `json:"v"`
	SecLevel     string `json:"l"`
	AuthProtocol string `json:"a"`
	AuthPassword string `json:"A"`
	SecName      string `json:"u"`
	PrivProtocol string `json:"x"`
	PrivPassword string `json:"X"`
	ContextName  string `json:"n"`
	//
	GoSnmpVersion gosnmp.SnmpVersion `json:"-"`
}

type HostsExported struct {
	content []byte
	data    map[string]HostProperties
}

// NewHostsExported is responsible reading kvData from the
// hosts-exported.json file.
func NewHostsExported() *HostsExported {
	var he HostsExported
	return &he
}

// GetHostProperties returns the properties of a host. It returns an
// error if the host is not found in the hosts-exported.json file.
func (he *HostsExported) GetHostProperties(host string) (HostProperties, error) {
	var err error
	var ret HostProperties
	var ok bool
	if he.data == nil {
		err = he.read()
		if err != nil {
			return ret, err
		}
		err = json.Unmarshal(he.content, &he.data)
		if err != nil {
			return ret, err
		}
	}
	if ret, ok = he.data[host]; !ok {
		return ret, ErrHostNotFound
	}
	switch ret.Version {
	case "1":
		ret.GoSnmpVersion = gosnmp.Version1
	case "3":
		ret.GoSnmpVersion = gosnmp.Version3
	default:
		ret.GoSnmpVersion = gosnmp.Version2c
	}
	ret.Host = host
	return ret, nil
}

type GoSnmpProperties struct {
	Version gosnmp.SnmpVersion
}

// read reads the hosts-exported.json file and stores
// the kvData in the HostsExported instance.
func (he *HostsExported) read() error {
	var err error
	he.content, err = os.ReadFile(f_hostsExportedFile)
	if err != nil {
		return err
	}
	return nil
}
