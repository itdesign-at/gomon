package v631

import (
	"errors"
	"strings"

	"github.com/nats-io/nats.go"
)

type NatsPublisher struct {
	To            string
	SubjectFields []string
}

func NewNatsPublisher(to string) *NatsPublisher {
	return &NatsPublisher{To: to}
}

func (p *NatsPublisher) WithSubject(subjectFields []string) *NatsPublisher {
	p.SubjectFields = subjectFields
	return p
}

func (p *NatsPublisher) Publish(data []byte) error {
	nc, err := nats.Connect(p.To)
	if err != nil {
		return err
	}
	if nc == nil {
		return errors.New("nats connection failed")
	}
	subject := strings.Join(p.SubjectFields, ".")
	// slog.Info("Publish", "subject", subject, "to", p.To)
	if err = nc.Publish(subject, data); err != nil {
		return err
	}
	nc.Close()
	return nil
}
