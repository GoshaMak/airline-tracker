package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

const group = "notifier-service"

type position struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Read      int    `json:"read"`
	Committed int64  `json:"committed"`
	End       int64  `json:"end"`
	Lag       int64  `json:"lag"`
}

func main() {
	if err := drain(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func drain() error {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	client, err := sarama.NewClient(strings.Split(os.Getenv("KAFKA_PEERS"), ","), config)
	if err != nil {
		return err
	}
	defer client.Close()
	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		return err
	}
	defer admin.Close()
	if err := waitForConsumerStop(ctx, admin); err != nil {
		return err
	}
	offsets, err := admin.ListConsumerGroupOffsets(group, nil)
	if err != nil {
		return err
	}
	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return err
	}
	defer consumer.Close()

	var positions []position
	for _, topic := range []string{os.Getenv("SUBSCRIPTION_CREATED_TOPIC"), os.Getenv("FLIGHT_UPDATED_TOPIC")} {
		partitions, err := client.Partitions(topic)
		if err != nil {
			return err
		}
		for _, partition := range partitions {
			p, err := drainPartition(ctx, client, consumer, topic, partition, offsets.GetBlock(topic, partition))
			if err != nil {
				return err
			}
			positions = append(positions, p)
		}
	}
	if err := commitAndVerify(admin, positions); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Group     string     `json:"group"`
		Positions []position `json:"positions"`
	}{group, positions})
}

func waitForConsumerStop(ctx context.Context, admin sarama.ClusterAdmin) error {
	for {
		groups, err := admin.DescribeConsumerGroups([]string{group})
		if err != nil {
			return err
		}
		if len(groups) == 1 && (groups[0].State == "Empty" || groups[0].State == "Dead") {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("consumer group did not stop: %w", ctx.Err())
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func drainPartition(
	ctx context.Context,
	client sarama.Client,
	consumer sarama.Consumer,
	topic string,
	partition int32,
	committed *sarama.OffsetFetchResponseBlock,
) (position, error) {
	p := position{Topic: topic, Partition: partition}
	start, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return p, err
	}
	p.End, err = client.GetOffset(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return p, err
	}
	if committed != nil && committed.Offset >= start {
		start = committed.Offset
	}
	if start > p.End {
		return p, fmt.Errorf("%s/%d: committed offset exceeds end", topic, partition)
	}
	if start == p.End {
		return p, nil
	}
	pc, err := consumer.ConsumePartition(topic, partition, start)
	if err != nil {
		return p, err
	}
	defer pc.Close()
	for start < p.End {
		select {
		case msg, ok := <-pc.Messages():
			if !ok {
				return p, fmt.Errorf("%s/%d: consumer closed before end", topic, partition)
			}
			start = msg.Offset + 1
			p.Read++
		case err := <-pc.Errors():
			if err == nil {
				return p, fmt.Errorf("%s/%d: consumer closed before end", topic, partition)
			}
			return p, err
		case <-ctx.Done():
			return p, ctx.Err()
		}
	}
	return p, nil
}

func commitAndVerify(admin sarama.ClusterAdmin, positions []position) error {
	commit := make(map[string]map[int32]sarama.OffsetAndMetadata)
	for _, p := range positions {
		if commit[p.Topic] == nil {
			commit[p.Topic] = make(map[int32]sarama.OffsetAndMetadata)
		}
		commit[p.Topic][p.Partition] = sarama.OffsetAndMetadata{Offset: p.End, Metadata: "ci-drain", LeaderEpoch: -1}
	}
	response, err := admin.AlterConsumerGroupOffsets(group, commit, nil)
	if err != nil {
		return err
	}
	for _, p := range positions {
		if code := response.Errors[p.Topic][p.Partition]; code != sarama.ErrNoError {
			return fmt.Errorf("commit %s/%d: %s", p.Topic, p.Partition, code)
		}
	}
	offsets, err := admin.ListConsumerGroupOffsets(group, nil)
	if err != nil {
		return err
	}
	for i := range positions {
		p := &positions[i]
		block := offsets.GetBlock(p.Topic, p.Partition)
		if block == nil {
			return fmt.Errorf("missing committed offset for %s/%d", p.Topic, p.Partition)
		}
		p.Committed, p.Lag = block.Offset, p.End-block.Offset
		if p.Lag != 0 {
			return fmt.Errorf("unread messages remain: %+v", p)
		}
	}
	return nil
}
