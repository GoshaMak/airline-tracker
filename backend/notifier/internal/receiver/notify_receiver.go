package receiver

import (
	"context"
	"fmt"
	"log/slog"
	"notifier/internal/receiver/handler"

	"github.com/IBM/sarama"
)

var (
	kafkaVersion = sarama.DefaultVersion.String()
	groupId      = "notifier-service"
)

type NotifyReceiver struct {
	cg       sarama.ConsumerGroup
	topics   []string
	groupId  string
	handlers map[string]handler.HandlerFunc
}

func NewNotifyReceiver(
	brokers,
	topics []string,
	handlers map[string]handler.HandlerFunc,
) (*NotifyReceiver, error) {
	cg, err := newNotificationConsumerGroup(brokers, groupId)
	if err != nil {
		return nil, err
	}

	return &NotifyReceiver{
		cg:       cg,
		topics:   topics,
		groupId:  groupId,
		handlers: handlers,
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
		handlers: n.handlers,
	}

	for {
		if err := n.cg.Consume(ctx, n.topics, handler); err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (n *NotifyReceiver) Close() error {
	if err := n.cg.Close(); err != nil {
		return err
	}
	return nil
}

func CloseConnection(ns *NotifyReceiver) error {
	return ns.Close()
}

type consumerGroupHandler struct {
	handlers map[string]handler.HandlerFunc
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
	const op = "consumerGroupHandler.ConsumeClaim"
	for msg := range claim.Messages() {
		handler, ok := h.handlers[msg.Topic]
		if !ok {
			slog.Warn(op, "unknown topic", msg.Topic, "msg", msg.Value)
			session.MarkMessage(msg, "")
			continue
		}
		slog.Debug(op, "handling message", msg)
		if err := handler(session.Context(), msg); err != nil {
			slog.Error(op, "err", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
