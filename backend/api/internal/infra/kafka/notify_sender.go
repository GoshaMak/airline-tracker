package kafka

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/samber/do/v2"
)

var kafkaVersion = sarama.DefaultVersion.String()

const producerShutdownTimeout = 10 * time.Second

type NotifySender struct {
	p               sarama.AsyncProducer
	sendGate        chan struct{}
	closing         chan struct{}
	closeSignalOnce sync.Once
	asyncCloseOnce  sync.Once
	done            chan struct{}
}

type deliveryResult struct {
	err error
}

func NewNotifySender(i do.Injector) (*NotifySender, error) {
	brokers := strings.Split(os.Getenv("KAFKA_PEERS"), ",")
	p, err := newNotificationProducer(brokers)
	if err != nil {
		return nil, err
	}
	return newNotifySender(p), nil
}

func newNotificationProducer(brokers []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	version, err := sarama.ParseKafkaVersion(kafkaVersion)
	if err != nil {
		return nil, err
	}
	config.Version = version
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 1
	config.Producer.Timeout = 3 * time.Second
	config.Net.DialTimeout = 3 * time.Second
	config.Net.ReadTimeout = 3 * time.Second
	config.Net.WriteTimeout = 3 * time.Second
	config.Metadata.Timeout = 3 * time.Second
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return sarama.NewAsyncProducer(brokers, config)
}

func newNotifySender(p sarama.AsyncProducer) *NotifySender {
	ns := &NotifySender{
		p:        p,
		sendGate: make(chan struct{}, 1),
		closing:  make(chan struct{}),
		done:     make(chan struct{}),
	}
	ns.sendGate <- struct{}{}
	go ns.dispatch()
	return ns
}

func (ns *NotifySender) dispatch() {
	defer close(ns.done)
	successes, failures := ns.p.Successes(), ns.p.Errors()
	for successes != nil || failures != nil {
		select {
		case message, ok := <-successes:
			if !ok {
				successes = nil
				continue
			}
			notifyResult(message, nil)
		case failure, ok := <-failures:
			if !ok {
				failures = nil
				continue
			}
			notifyResult(failure.Msg, failure.Err)
		}
	}
}

func notifyResult(message *sarama.ProducerMessage, err error) {
	if message == nil {
		return
	}
	if result, ok := message.Metadata.(chan deliveryResult); ok {
		result <- deliveryResult{err: err}
	}
}

// SendContext submits one message and waits for its broker result. The one
// shared dispatcher drains Sarama responses, including late responses after a
// caller deadline; no goroutine is created per send.
func (ns *NotifySender) SendContext(
	ctx context.Context,
	topic,
	key,
	eventID string,
	version int,
	carrier map[string]string,
	msg []byte,
) error {
	const op = "NotifySender.SendContext"
	if err := ctx.Err(); err != nil {
		return err
	}
	if topic == "" || eventID == "" || version < 1 {
		return fmt.Errorf("%s: topic, event_id and positive schema version are required", op)
	}
	result := make(chan deliveryResult, 1)
	message := &sarama.ProducerMessage{
		Topic:    topic,
		Key:      sarama.StringEncoder(key),
		Value:    sarama.ByteEncoder(msg),
		Metadata: result,
	}
	message.Headers = append(message.Headers,
		sarama.RecordHeader{
			Key:   []byte("event_id"),
			Value: []byte(eventID),
		},
		sarama.RecordHeader{
			Key:   []byte("schema_version"),
			Value: fmt.Append(nil, version),
		},
	)
	if correlationID := carrier["x-correlation-id"]; correlationID != "" {
		message.Headers = append(message.Headers, sarama.RecordHeader{
			Key:   []byte("correlation_id"),
			Value: []byte(correlationID),
		})
	}
	keys := make([]string, 0, len(carrier))
	for k := range carrier {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		message.Headers = append(message.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(carrier[k]),
		})
	}

	select {
	case <-ns.sendGate:
	case <-ns.closing:
		return fmt.Errorf("%s: producer is closed", op)
	case <-ctx.Done():
		return ctx.Err()
	}
	gateHeld := true
	defer func() {
		if gateHeld {
			ns.sendGate <- struct{}{}
		}
	}()
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-ns.closing:
		return fmt.Errorf("%s: producer is closed", op)
	default:
	}
	select {
	case ns.p.Input() <- message:
	case <-ns.closing:
		return fmt.Errorf("%s: producer is closed", op)
	case <-ctx.Done():
		return ctx.Err()
	}
	ns.sendGate <- struct{}{}
	gateHeld = false
	select {
	case outcome := <-result:
		if outcome.err != nil {
			return fmt.Errorf("%s: %w", op, outcome.err)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.closing:
		return fmt.Errorf("%s: producer is closing", op)
	}
}

func (ns *NotifySender) Close() error {
	const op = "NotifySender.Close"
	timer := time.NewTimer(producerShutdownTimeout)
	defer timer.Stop()
	ns.closeSignalOnce.Do(func() { close(ns.closing) })
	select {
	case <-ns.sendGate:
		ns.asyncCloseOnce.Do(ns.p.AsyncClose)
		ns.sendGate <- struct{}{}
	case <-timer.C:
		return fmt.Errorf("%s: waiting for active send exceeded %s", op, producerShutdownTimeout)
	}
	select {
	case <-ns.done:
		return nil
	case <-timer.C:
		return fmt.Errorf("%s: producer drain exceeded %s", op, producerShutdownTimeout)
	}
}

func CloseConnection(ns *NotifySender) error {
	return ns.Close()
}
