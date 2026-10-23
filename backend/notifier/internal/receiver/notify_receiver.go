package receiver

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"notifier/internal/receiver/handler"
	"sync/atomic"
	"time"

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
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = time.Second
	config.Consumer.Group.Session.Timeout = 30 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Net.DialTimeout = 3 * time.Second
	config.Net.ReadTimeout = 5 * time.Second
	config.Net.WriteTimeout = 5 * time.Second
	config.Consumer.Retry.Backoff = time.Second

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return sarama.NewConsumerGroup(brokers, groupId, config)
}

func (n *NotifyReceiver) StartConsuming(ctx context.Context) error {
	handler := &consumerGroupHandler{
		handlers: n.handlers,
	}

	backoff := time.Second
	for {
		err := n.cg.Consume(ctx, n.topics, handler)
		failed := handler.failed.Swap(false)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err == nil && !failed {
			backoff = time.Second
			continue
		}
		if err != nil {
			slog.ErrorContext(ctx, "Kafka consume failed; retrying", "err", err, "backoff", backoff)
		} else {
			slog.WarnContext(ctx, "Kafka message handler failed; retrying consumer", "backoff", backoff)
		}
		delay := backoff + time.Duration(rand.Int63n(int64(backoff/4+1)))
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
		backoff = min(backoff*2, 30*time.Second)
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
	failed   atomic.Bool
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
			err := fmt.Errorf("%s: no handler for Kafka topic", op)
			h.failed.Store(true)
			slog.Error(op, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset, "err", err)
			return err
		}
		if err := handler(session.Context(), msg); err != nil {
			h.failed.Store(true)
			slog.Error(op, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset, "err", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		session.MarkMessage(msg, "")
		slog.Debug("message handled", "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)
	}
	return nil
}
