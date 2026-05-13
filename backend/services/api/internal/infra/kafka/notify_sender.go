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
	// cg      sarama.ConsumerGroup
	topic string
	// groupId string
	// msgs    chan string
}

func NewNotifySender(i do.Injector) (*NotifySender, error) {
	brokers := strings.Split(os.Getenv("KAFKA_PEERS"), ",")
	p, err := newNotificationProducer(brokers)
	if err != nil {
		return nil, err
	}

	topic := "flights"
	// groupId := topic + "-group"

	// cg, err := newNotificationConsumerGroup(brokers, groupId)
	// if err != nil {
	// 	p.Close()
	// 	return nil, err
	// }

	return &NotifySender{
		p: p,
		// cg:      cg,
		topic: topic,
		// groupId: groupId,
		// msgs:    make(chan string),
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

// func newNotificationConsumerGroup(brokers []string, groupId string) (sarama.ConsumerGroup, error) {
// 	config := sarama.NewConfig()
//
// 	version, err := sarama.ParseKafkaVersion(kafkaVersion)
// 	if err != nil {
// 		return nil, err
// 	}
// 	config.Version = version
// 	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
// 		sarama.NewBalanceStrategyRange(),
// 	}
// 	config.Consumer.Offsets.Initial = sarama.OffsetNewest
// 	config.Consumer.Offsets.AutoCommit.Enable = true
//
// 	if err := config.Validate(); err != nil {
// 		return nil, err
// 	}
//
// 	return sarama.NewConsumerGroup(brokers, groupId, config)
// }

func (ns *NotifySender) SendMessage(msg string) error {
	m := &sarama.ProducerMessage{
		Topic: ns.topic,
		Value: sarama.StringEncoder(msg),
	}
	partition, offset, err := ns.p.SendMessage(m)
	if err != nil {
		return err
	}

	slog.Debug(fmt.Sprintf("message sent: topic=%s partition=%d offset=%d", ns.topic, partition, offset))
	return nil
}

// func (n *SendNotifier) StartConsuming(ctx context.Context) error {
// 	handler := &consumerGroupHandler{
// 		msgs: n.msgs,
// 	}
//
// 	for {
// 		if err := n.cg.Consume(ctx, []string{n.topic}, handler); err != nil {
// 			return err
// 		}
//
// 		if ctx.Err() != nil {
// 			return ctx.Err()
// 		}
// 	}
// }

// func (n *SendNotifier) ReadMessage(ctx context.Context) (string, error) {
// 	select {
// 	case msg := <-n.msgs:
// 		return msg, nil
// 	case <-ctx.Done():
// 		return "", ctx.Err()
// 	}
// }

func (ns *NotifySender) Close() error {
	var err error
	if e := ns.p.Close(); e != nil {
		err = e
	}
	// if e := n.cg.Close(); e != nil {
	// 	err = errors.Join(err, e)
	// }
	return err
}

func CloseConnection(ns *NotifySender) error {
	return ns.Close()
}

// type consumerGroupHandler struct {
// 	msgs chan<- string
// }
//
// func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
// 	return nil
// }
//
// func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
// 	return nil
// }
//
// func (h *consumerGroupHandler) ConsumeClaim(
// 	session sarama.ConsumerGroupSession,
// 	claim sarama.ConsumerGroupClaim,
// ) error {
// 	for msg := range claim.Messages() {
// 		h.msgs <- string(msg.Value)
//
// 		session.MarkMessage(msg, "")
// 	}
//
// 	return nil
// }
