package mongo

import (
	"context"
	"fmt"
	"notifier/internal/receiver/domain"
	"notifier/internal/receiver/domain/repository"
	"notifier/internal/receiver/infra/mongo/model"

	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type notificationRepository struct {
	db *mongo.Database
}

func NewNotificationRepository(i do.Injector) (repository.NotificationRepository, error) {
	db := do.MustInvoke[*mongo.Database](i)

	collection := db.Collection("notifications")
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	}
	collection.Indexes().CreateOne(context.Background(), indexModel)

	return &notificationRepository{
		db: db,
	}, nil
}

func (r *notificationRepository) Save(ctx context.Context, n domain.Notification) error {
	const op = "NotificationRepository.Save"

	doc := model.ToNotificationDoc(n)
	collection := r.db.Collection("notifications")

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *notificationRepository) ListNotSent(ctx context.Context) ([]domain.Notification, error) {
	const op = "NotificationRepository.ListNotSent"

	collection := r.db.Collection("notifications")

	filter := bson.M{"status": bson.M{"$ne": domain.NotificationSent.String()}}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer cursor.Close(ctx)

	var docs []model.NotificationDoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	notifications := make([]domain.Notification, 0, len(docs))
	for _, doc := range docs {
		n, err := model.FromNotificationDoc(doc)
		if err != nil {
			continue
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (r *notificationRepository) Mark(ctx context.Context, n domain.Notification, newStatus domain.NotificationStatus) error {
	const op = "NotificationRepository.Mark"

	collection := r.db.Collection("notifications")
	filter := bson.M{"_id": n.Id.String()}
	update := bson.M{"$set": bson.M{"status": newStatus.String()}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("%s: notification not found", op)
	}

	return nil
}
