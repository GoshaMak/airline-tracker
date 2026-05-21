package kafka

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/IBM/sarama"
	"github.com/samber/do/v2"
)

var (
	kafkaVersion = sarama.DefaultVersion.String()
)

type NotifySender struct {
	p sarama.SyncProducer
}

func NewNotifySender(i do.Injector) (*NotifySender, error) {
	brokers := strings.Split(os.Getenv("KAFKA_PEERS"), ",")
	p, err := newNotificationProducer(brokers)
	if err != nil {
		return nil, err
	}

	return &NotifySender{
		p: p,
	}, nil
}

func newNotificationProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(kafkaVersion)
	if err != nil {
		return nil, err
	}
	config.Version = version
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return sarama.NewSyncProducer(brokers, config)
}

func (ns *NotifySender) SendMessage(topic string, msg []byte) error {
	const op = "NotifySender.SendMessage"
	m := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
	}
	slog.Debug(op, "msg", string(msg))
	partition, offset, err := ns.p.SendMessage(m)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	slog.Debug(op+": msg meta", "topic", topic, "partition", partition, "offset", offset)
	return nil
}

func (ns *NotifySender) Close() error {
	var err error
	if e := ns.p.Close(); e != nil {
		err = e
	}
	return err
}

func CloseConnection(ns *NotifySender) error {
	return ns.Close()
}
