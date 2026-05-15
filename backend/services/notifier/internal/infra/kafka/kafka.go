package kafka

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/IBM/sarama"
)

var (
	kafkaVersion = sarama.DefaultVersion.String()
)

type NotifyReceiver struct {
	cg      sarama.ConsumerGroup
	topic   string
	groupId string
	msgs    chan []byte
}

func NewNotifyReceiver(topic string) (*NotifyReceiver, error) {
	brokers := strings.Split(os.Getenv("KAFKA_PEERS"), ",")

	groupId := topic + "-group"

	cg, err := newNotificationConsumerGroup(brokers, groupId)
	if err != nil {
		return nil, err
	}

	return &NotifyReceiver{
		cg:      cg,
		topic:   topic,
		groupId: groupId,
		msgs:    make(chan []byte),
	}, nil
}

func newNotificationConsumerGroup(
	brokers []string,
	groupId string,
) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(kafkaVersion)
	if err != nil {
		return nil, err
	}
	config.Version = version
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRange(), // TODO: choose strategy wisely
	}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = true

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return sarama.NewConsumerGroup(brokers, groupId, config)
}

func (n *NotifyReceiver) StartConsuming(ctx context.Context) error {
	handler := &consumerGroupHandler{
		msgs: n.msgs,
	}

	for {
		if err := n.cg.Consume(ctx, []string{n.topic}, handler); err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (n *NotifyReceiver) ReadMessage(ctx context.Context) ([]byte, error) {
	select {
	case msg := <-n.msgs:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (n *NotifyReceiver) Close() error {
	var err error
	if e := n.cg.Close(); e != nil {
		err = errors.Join(err, e)
	}
	return err
}

func CloseConnection(ns *NotifyReceiver) error {
	return ns.Close()
}

type consumerGroupHandler struct {
	msgs chan<- []byte
}

func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for msg := range claim.Messages() {
		h.msgs <- msg.Value

		session.MarkMessage(msg, "")
	}

	return nil
}
