package db

import (
	"context"
	"fmt"
	"imgpipe/pkg/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type JobRepository struct {
	db *mongo.Database
}

func NewJobRepository(db *mongo.Database) *JobRepository {
	return &JobRepository{
		db: db,
	}
}

func (r *JobRepository) CreateJob(ctx context.Context, job *models.Job) error {
	if job.ID == "" {
		job.ID = bson.NewObjectID().Hex()
	}
	job.CreatedAt = time.Now()
	job.Status = "pending"

	coll := r.db.Collection("jobs")

	_, err := coll.InsertOne(ctx, job)
	if err != nil {
		return fmt.Errorf("can't save job to db: %w", err)
	}

	return nil
}

func (r *JobRepository) GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	filter := bson.M{"_id": id}

	var job models.Job

	err := r.db.Collection("jobs").FindOne(ctx, filter).Decode(&job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *JobRepository) UpdateJobStatus(ctx context.Context, id string, status string, resultURL string, errMsg string) error {
	filter := bson.M{"_id": id}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"result_url": resultURL,
			"error":      errMsg,
		},
	}

	_, err := r.db.Collection("jobs").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("can't update job %s: %w", id, err)
	}

	return nil
}
