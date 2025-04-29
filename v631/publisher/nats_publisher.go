package publisher

import (
	"encoding/json"
	"errors"
	"github.com/itdesign-at/golib/keyvalue"

	"github.com/nats-io/nats.go"
)

type NatsPublisher struct {
	// args are used to configure the publisher
	args keyvalue.Record

	// debug is used to enable debug logging to STDERR with slog
	debug bool

	// subject is the NATS subject
	subject string

	// to is the NATS server address
	to string
}

func NewNatsPublisher(to string) *NatsPublisher {
	var np NatsPublisher
	np.args = keyvalue.NewRecord()
	np.to = to
	return &np
}

func (p *NatsPublisher) WithSubject(subject string) *NatsPublisher {
	p.subject = subject
	return p
}

func (p *NatsPublisher) WithArgs(args keyvalue.Record) *NatsPublisher {
	p.args = args
	if args.Exists("Debug") {
		p.debug = true
	}
	return p
}

// PublishJson publishes data as JSON to the NATS server.
func (p *NatsPublisher) PublishJson(data any) error {
	j, e := json.Marshal(data)
	if e != nil {
		return e
	}
	return p.Publish(j)
}

// Publish publishes data to the NATS server.
func (p *NatsPublisher) Publish(data []byte) error {
	nc, err := nats.Connect(p.to)
	if err != nil {
		return err
	}
	if nc == nil {
		return errors.New("nats connection failed")
	}
	if err = nc.Publish(p.subject, data); err != nil {
		return err
	}
	nc.Close()
	return nil
}
