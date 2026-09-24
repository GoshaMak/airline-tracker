package usecase

import (
	"api/internal/publisher/domain"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// outboxBuilder supplies valid objects while keeping each test's change local.
type outboxBuilder struct{ item domain.Outbox }

func newOutboxBuilder() outboxBuilder {
	return outboxBuilder{domain.Outbox{
		Id: uuid.New(), Topic: "flights", Payload: &SendPayload{data: []byte(`{"flight":"updated"}`)},
		CreatedAt: time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC),
	}}
}
func (b outboxBuilder) withPayload(p domain.Payload) outboxBuilder { b.item.Payload = p; return b }
func (b outboxBuilder) build() domain.Outbox                       { return b.item }

type memoryOutbox struct {
	items   []domain.Outbox
	listErr error
	markErr error
	marked  int
}

func (r *memoryOutbox) Save(_ context.Context, item domain.Outbox) error {
	r.items = append(r.items, item)
	return nil
}
func (r *memoryOutbox) ListNotSent(_ context.Context, _ func() domain.Payload) ([]domain.Outbox, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.items, nil
}
func (r *memoryOutbox) MarkAsSent(_ context.Context, item domain.Outbox) error {
	r.marked++
	if r.markErr != nil {
		return r.markErr
	}
	for i := range r.items {
		if r.items[i].Id == item.Id {
			r.items[i] = item
			return nil
		}
	}
	return errors.New("outbox item missing")
}

type senderSpy struct {
	calls   int
	topic   string
	payload []byte
	err     error
}

func (s *senderSpy) SendMessage(topic string, payload []byte) error {
	s.calls++
	s.topic, s.payload = topic, append([]byte(nil), payload...)
	return s.err
}

type failingPayload struct{ err error }

func (p *failingPayload) MarshalJSON() ([]byte, error) { return nil, p.err }
func (p *failingPayload) UnmarshalJSON([]byte) error   { return nil }

func TestPublisherUsecase_Publish(t *testing.T) {
	t.Run("positive classic: publishes and marks stored item", func(t *testing.T) {
		// Arrange
		item := newOutboxBuilder().build()
		repo := &memoryOutbox{items: []domain.Outbox{item}}
		sender := &senderSpy{}
		uc := &PublisherUsecase{repo: repo, ns: sender}
		// Act
		err := uc.Publish(context.Background())
		// Assert
		if err != nil || repo.items[0].SentAt == nil || sender.topic != item.Topic || string(sender.payload) != `{"flight":"updated"}` {
			t.Fatalf("err=%v, item=%+v, sender=%+v", err, repo.items[0], sender)
		}
	})
	t.Run("negative classic: invalid payload is not sent", func(t *testing.T) {
		// Arrange
		failure := errors.New("encoding failed")
		repo := &memoryOutbox{items: []domain.Outbox{newOutboxBuilder().withPayload(&failingPayload{err: failure}).build()}}
		sender := &senderSpy{}
		uc := &PublisherUsecase{repo: repo, ns: sender}
		// Act
		err := uc.Publish(context.Background())
		// Assert
		if !errors.Is(err, failure) || sender.calls != 0 || repo.marked != 0 {
			t.Fatalf("err=%v, sender=%+v, marked=%d", err, sender, repo.marked)
		}
	})
	t.Run("positive London: sends once then marks once", func(t *testing.T) {
		// Arrange
		repo := &memoryOutbox{items: []domain.Outbox{newOutboxBuilder().build()}}
		sender := &senderSpy{}
		uc := &PublisherUsecase{repo: repo, ns: sender}
		// Act
		err := uc.Publish(context.Background())
		// Assert
		if err != nil || sender.calls != 1 || repo.marked != 1 {
			t.Fatalf("err=%v, calls=%d, marked=%d", err, sender.calls, repo.marked)
		}
	})
	t.Run("negative London: sender failure prevents mark", func(t *testing.T) {
		// Arrange
		failure := errors.New("broker unavailable")
		repo := &memoryOutbox{items: []domain.Outbox{newOutboxBuilder().build()}}
		sender := &senderSpy{err: failure}
		uc := &PublisherUsecase{repo: repo, ns: sender}
		// Act
		err := uc.Publish(context.Background())
		// Assert
		if !errors.Is(err, failure) || sender.calls != 1 || repo.marked != 0 {
			t.Fatalf("err=%v, calls=%d, marked=%d", err, sender.calls, repo.marked)
		}
	})
	t.Run("negative: repository list error", func(t *testing.T) {
		// Arrange
		failure := errors.New("database unavailable")
		repo := &memoryOutbox{listErr: failure}
		sender := &senderSpy{}
		uc := &PublisherUsecase{repo: repo, ns: sender}
		// Act
		err := uc.Publish(context.Background())
		// Assert
		if !errors.Is(err, failure) || sender.calls != 0 {
			t.Fatalf("err=%v, calls=%d", err, sender.calls)
		}
	})
}

func TestSendPayload_MarshalJSON(t *testing.T) {
	// Arrange
	p := &SendPayload{data: []byte(`{"ok":true}`)}
	// Act
	got, err := p.MarshalJSON()
	// Assert
	if err != nil || string(got) != `{"ok":true}` {
		t.Fatalf("got=%q, err=%v", got, err)
	}
}

func TestSendPayload_MarshalJSON_Invalid(t *testing.T) {
	// Arrange
	p := &SendPayload{data: []byte("broken")}
	// Act
	_, err := p.MarshalJSON()
	// Assert
	if err == nil {
		t.Fatal("invalid JSON accepted")
	}
}

func TestSendPayload_UnmarshalJSON(t *testing.T) {
	// Arrange
	input := []byte(`{"ok":true}`)
	p := &SendPayload{}
	// Act
	err := p.UnmarshalJSON(input)
	input[0] = 'x'
	// Assert
	if err != nil || string(p.data) != `{"ok":true}` {
		t.Fatalf("data=%q, err=%v", p.data, err)
	}
}

func TestSendPayload_UnmarshalJSON_Invalid(t *testing.T) {
	// Arrange
	p := &SendPayload{data: []byte(`{"previous":true}`)}
	// Act
	err := p.UnmarshalJSON([]byte("broken"))
	// Assert
	if err == nil || string(p.data) != `{"previous":true}` {
		t.Fatalf("err=%v, data=%q", err, p.data)
	}
}
