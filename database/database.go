package database

import (
	"context"
	"log"
	"time"

	"github.com/makespositivechange/mongo-graphql-api/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var connectionString string = "mongodb://localhost:27017"

type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	clientOptions := options.Client().ApplyURI(connectionString)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error while connecting with mongo DB:%v", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Error while pinging mongoDB:%v", err)
	}

	return &DB{
		client: client,
	}
}

func (db *DB) GetJob(id string) *model.Joblisting {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	var joblisting model.Joblisting
	err := jobCollec.FindOne(ctx, filter).Decode(&joblisting)
	if err != nil {
		log.Printf("Could not find record for jobid:%s and error is:%v", id, err)
		return nil
	}
	return &joblisting
}

func (db *DB) GetJobs() []*model.Joblisting {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := jobCollec.Find(ctx, bson.D{})
	if err != nil {
		log.Fatalf("Failed to fetch jobs from database:%v", err)
	}
	var joblistings []*model.Joblisting
	err = cursor.All(ctx, &joblistings)
	if err != nil {
		panic(err)
	}
	return joblistings
}

func (db *DB) CreateJobListing(joblist model.CreateJobListingInput) *model.Joblisting {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	inserted, err := jobCollec.InsertOne(ctx, bson.M{
		"title":       joblist.Title,
		"url":         joblist.URL,
		"description": joblist.Description,
		"company":     joblist.Company,
	})

	if err != nil {
		log.Fatalf("Failed to insert record in database:%v", err)
	}
	insertedID := inserted.InsertedID.(primitive.ObjectID).Hex()
	log.Println("insertedID", insertedID)
	returnJobListing := model.Joblisting{ID: insertedID,
		Title:       joblist.Title,
		URL:         joblist.URL,
		Description: joblist.Description,
		Company:     joblist.Company}
	return &returnJobListing
}

func (db *DB) UpdateJobListing(jobid string, jobInfo model.UpdateJobListingInput) *model.Joblisting {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	updateJoblist := bson.M{}

	if jobInfo.Title != nil {
		updateJoblist["title"] = jobInfo.Title
	}

	if jobInfo.Description != nil {
		updateJoblist["description"] = jobInfo.Description
	}

	if jobInfo.URL != nil {
		updateJoblist["url"] = jobInfo.URL
	}
	_id, _ := primitive.ObjectIDFromHex(jobid)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": updateJoblist}

	result := jobCollec.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))

	var joblisting model.Joblisting

	if err := result.Decode(&joblisting); err != nil {
		log.Fatal(err)
	}
	return &joblisting
}

func (db *DB) DeleteJobListing(jobid string) *model.DeletedJobResponse {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(jobid)
	filter := bson.M{"_id": _id}
	_, err := jobCollec.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	return &model.DeletedJobResponse{DeletedJobID: jobid}
}
