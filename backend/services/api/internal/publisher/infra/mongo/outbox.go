package mongo

import (
	"api/internal/publisher/domain"
	"api/internal/publisher/domain/repository"
	"api/internal/publisher/infra/mongo/model"
	"context"
	"fmt"

	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type outboxRepository struct {
	db *mongo.Database
}

func NewOutboxRepository(i do.Injector) (repository.OutboxRepository, error) {
	db := do.MustInvoke[*mongo.Database](i)

	collection := db.Collection("outbox")
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "sent_at", Value: 1}},
	}
	collection.Indexes().CreateOne(context.Background(), indexModel)

	return &outboxRepository{
		db: db,
	}, nil
}

func (r *outboxRepository) Save(ctx context.Context, ob domain.Outbox) error {
	const op = "OutboxRepository.Save"

	doc, err := model.ToOutboxDoc(ob)
	if err != nil {
		return fmt.Errorf("%s: marshal error: %w", op, err)
	}

	collection := r.db.Collection("outbox")
	_, err = collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *outboxRepository) ListNotSent(ctx context.Context, newPayload func() domain.Payload) ([]domain.Outbox, error) {
	const op = "OutboxRepository.ListNotSent"

	collection := r.db.Collection("outbox")

	filter := bson.M{"sent_at": nil}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer cursor.Close(ctx)

	var docs []model.OutboxDoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	outboxes := make([]domain.Outbox, 0, len(docs))
	for _, doc := range docs {
		payloadInstance := newPayload()

		ob, err := model.FromOutboxDoc(doc, payloadInstance)
		if err != nil {
			fmt.Printf("Error converting outbox doc %s: %v\n", doc.ID, err)
			continue
		}
		outboxes = append(outboxes, ob)
	}

	return outboxes, nil
}

func (r *outboxRepository) MarkAsSent(ctx context.Context, ob domain.Outbox) error {
	const op = "OutboxRepository.MarkAsSent"

	collection := r.db.Collection("outbox")
	filter := bson.M{"_id": ob.Id.String()}

	update := bson.M{"$set": bson.M{"sent_at": ob.SentAt}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("%s: outbox message not found", op)
	}

	return nil
}
