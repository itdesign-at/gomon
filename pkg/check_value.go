package pkg

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/itdesign-at/golib/converter"
	"github.com/itdesign-at/golib/keyvalue"
)

const (
	Exit    = "Exit"
	Host    = "Host"
	Node    = "Node"
	Service = "Service"
	State   = "State"
	Type    = "Type"
)

var validStates = []string{
	"OK", "WARNING", "CRITICAL", "UNKNOWN",
}

type CheckValue struct {
	args keyvalue.Record
}

func NewCheckValue(args keyvalue.Record) *CheckValue {
	var cv CheckValue
	cv.args = keyvalue.NewRecord()
	cv.add(args)
	return &cv
}

func (cv *CheckValue) Add(data any) {
	switch v := data.(type) {
	case keyvalue.Record:
		cv.add(v)
	case map[string]any:
		cv.add(v)
	case string:
		var tmp keyvalue.Record
		_ = json.Unmarshal([]byte(v), &tmp)
		cv.add(tmp)
	}
}

// GetExit returns the exit code based
// on the "Exit" key or on the "State" key.
// It returns 0 when both are not set.
func (cv *CheckValue) GetExit() int {
	if cv.args.Exists(Exit) {
		return cv.args.Int(Exit, true)
	}
	if state := cv.args.String(State); state != "" {
		for i, str := range validStates {
			if state == str {
				return i
			}
		}
		return 3
	}
	return 0
}

// PrepareExit sets both "Exit" and "State"
func (cv *CheckValue) PrepareExit() {
	exit := cv.GetExit()
	// check 0, 1, 2 (3 could be converted to CRITICAL)
	if exit > -1 && exit < 3 {
		cv.args[Exit] = exit
		cv.args[State] = validStates[exit]
		return
	}
	if exit < 0 {
		goto exit3
	}
	if convert := cv.args.String("convert", true); convert != "" {
		switch convert {
		case "1":
		case "UNKNOWN CRITICAL":
			cv.args[Exit] = 2
			cv.args[State] = validStates[2]
			return
		}
	}
exit3:
	cv.args[Exit] = 3
	cv.args[State] = validStates[3]
	return
}

func (cv *CheckValue) GetMacros() keyvalue.Record {
	return cv.args
}

// GetData returns key/value records with keys starting with
// an upper case letter.
func (cv *CheckValue) GetData() keyvalue.Record {
	var kv = keyvalue.NewRecord()
	for key, value := range cv.args {
		if unicode.IsUpper(rune(key[0])) {
			kv[key] = value
		}
	}
	return kv
}

func (cv *CheckValue) Bye() int {
	cv.PrepareExit()
	md, err := GetMetricDefinition(cv.args.String(Type))
	if err != nil {
		panic(err)
	}
	if md == nil {
		panic("is nil")
	}

	metric := cv.args.String(Type)
	node := cv.node(true)
	host := converter.Normalize(cv.args.String(Host))
	service := converter.Normalize(cv.args.String(Service))

	// in most cases to = localhost (default port 4222)
	to := md.String("$to")
	publisher := NewNatsPublisher(to).WithSubject(
		[]string{"V2", metric, node, host, service},
	)

	data, _ := json.Marshal(cv.GetData())
	err = publisher.Publish(data)
	if err != nil {
		slog.Error("CheckValue: publish failed", err)
		return 3
	}

	nhs := node + "/" + host + "/" + service
	from := strings.Replace(md.String("$from"), "{{.NHS}}", nhs, -1)
	from = strings.Replace(from, "{{.K}}", cv.args.String(Type), -1)

	type tmp struct {
		DSN  string `json:"DSN"`
		Text string `json:"Text"`
	}

	var t tmp
	t.DSN = from
	t.Text = fmt.Sprintf("%.2fms rtt, %d%% packet loss",
		1000*cv.args.Float64("Rtt", true), cv.args.Int("Pl", true))

	out, _ := json.Marshal(t)
	fmt.Println(string(out))

	return cv.args.Int(Exit)
}

// node returns the node name or sets it and returns it
func (cv *CheckValue) node(shouldNormalize bool) string {
	var node string
	node = cv.args.String(Node)
	if node == "" {
		node, _ = os.Hostname()
		cv.args[Node] = node
	}
	if shouldNormalize {
		return converter.Normalize(node)
	}
	return node
}

func (cv *CheckValue) add(data keyvalue.Record) {
	if data.Exists(State) && data.Exists(Exit) {
		slog.Error("CheckValue: both State and Exit are set")
		return
	}
	for key, value := range data {
		switch key {
		case "k":
			cv.args[Type] = data.String(key, true)
		case "h":
			cv.args[Host] = data.String(key, true)
		case "s":
			cv.args[Service] = data.String(key, true)
		case "n":
			cv.args[Node] = data.String(key, true)
		case Exit:
			// default = UNKNOWN
			cv.args[Exit] = 3
			switch v := data[Exit].(type) {
			case int: // 0, 1, 2, 3
				if v >= 0 && v <= 3 {
					cv.args[Exit] = v
				}
			case string: // "0", "1", "2", "3"
				idx, err := strconv.Atoi(v)
				if err == nil && idx >= 0 && idx <= 3 {
					cv.args[Exit] = idx
				}
			}
		case State:
			// default = UNKNOWN
			cv.args[State] = validStates[3]
			switch v := data[State].(type) {
			case string:
				if slices.Contains(validStates, v) {
					cv.args[State] = v
				}
			}
		default:
			cv.args[key] = value
		}
	}
}
