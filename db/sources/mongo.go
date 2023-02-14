package sources

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient interface {
	Aggregate(ctx context.Context, table string, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error)
	UpsertOne(ctx context.Context, table string, filter interface{}, update interface{}) error
	UpsertMany(ctx context.Context, table string, filter interface{}, update interface{}) error
	InsertOne(ctx context.Context, table string, document interface{}, opts ...*options.InsertOneOptions) error
	InsertMany(ctx context.Context, table string, documents []interface{}, opts ...*options.InsertManyOptions) error
	Find(ctx context.Context, table string, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error)
	FindOne(ctx context.Context, table string, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	DeleteOne(ctx context.Context, table string, filter interface{}, opts ...*options.DeleteOptions) error
	DeleteMany(ctx context.Context, table string, filter interface{}, opts ...*options.DeleteOptions) error
	UpdateOne(ctx context.Context, table string, filter interface{}, update interface{}, opts ...*options.UpdateOptions) error
	DoesExist(ctx context.Context, table string, filter bson.D, opts ...*options.FindOneOptions) (bool, error)
	CreateIndex(ctx context.Context, table string, keys ...bson.E) (string, error)
	Count(ctx context.Context, table string, filter interface{}, opts ...*options.CountOptions) (int64, error)
	Disconnect(ctx context.Context) error
	WithSession() (MongoClient, error)
	WithTransaction(ctx context.Context, callback func(sessCtx mongo.SessionContext) (interface{}, error)) (interface{}, error)
}

type mongoClient struct {
	cl      *mongo.Client
	db      *mongo.Database
	session mongo.Session
}

func NewMongoClient(ctx context.Context, uri, dbName string) MongoClient {
	opts := options.Client()
	opts.ApplyURI(uri)
	opts.SetMaxPoolSize(5)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatal("Error while connecting to MongoClient", err)
		panic(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Error while pinging MongoClient", err)
		panic(err)
	}

	db := client.Database(dbName)

	return &mongoClient{
		db: db,
		cl: client,
	}
}

func (mc *mongoClient) WithSession() (MongoClient, error) {
	session, err := mc.cl.StartSession()
	if err != nil {
		return nil, err
	}

	return &mongoClient{
		db:      mc.db,
		cl:      mc.cl,
		session: session,
	}, nil
}

func (mc *mongoClient) WithTransaction(
	ctx context.Context,
	callback func(sessCtx mongo.SessionContext) (interface{}, error),
) (interface{}, error) {
	if mc.session == nil {
		return nil, fmt.Errorf("empty session")
	}

	return mc.session.WithTransaction(ctx, callback)
}

func (mc *mongoClient) getCollection(table string) *mongo.Collection {
	return mc.db.Collection(table)
}

func (mc *mongoClient) Aggregate(ctx context.Context, table string, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	coll := mc.getCollection(table)

	return coll.Aggregate(ctx, pipeline, opts...)
}

func (mc *mongoClient) Count(ctx context.Context, table string, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	coll := mc.getCollection(table)

	return coll.CountDocuments(ctx, filter, opts...)
}

func (mc *mongoClient) UpsertOne(ctx context.Context, table string, filter interface{}, update interface{}) error {
	coll := mc.getCollection(table)

	_, err := coll.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

func (mc *mongoClient) UpsertMany(ctx context.Context, table string, filter interface{}, update interface{}) error {
	coll := mc.getCollection(table)

	_, err := coll.UpdateMany(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

func (mc *mongoClient) Disconnect(ctx context.Context) error {
	return mc.cl.Disconnect(ctx)
}

func (mc *mongoClient) UpdateOne(ctx context.Context, table string, filter interface{}, update interface{}, opts ...*options.UpdateOptions) error {
	coll := mc.getCollection(table)

	_, err := coll.UpdateOne(ctx, filter, update, opts...)

	return err
}

func (mc *mongoClient) InsertOne(ctx context.Context, table string, document interface{}, opts ...*options.InsertOneOptions) error {
	coll := mc.getCollection(table)

	_, err := coll.InsertOne(ctx, document, opts...)

	return err
}

func (mc *mongoClient) InsertMany(ctx context.Context, table string, documents []interface{}, opts ...*options.InsertManyOptions) error {
	coll := mc.getCollection(table)

	_, err := coll.InsertMany(ctx, documents, opts...)

	return err
}

func (mc *mongoClient) Find(ctx context.Context, table string, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error) {
	coll := mc.db.Collection(table)

	return coll.Find(ctx, filter, opts...)
}

func (mc *mongoClient) FindOne(ctx context.Context, table string, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	coll := mc.db.Collection(table)

	return coll.FindOne(ctx, filter, opts...)
}

func (mc *mongoClient) DoesExist(ctx context.Context, table string, filter bson.D, opts ...*options.FindOneOptions) (bool, error) {
	result := make(bson.M)

	coll := mc.db.Collection(table)

	err := coll.FindOne(ctx, filter, opts...).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (mc *mongoClient) CreateIndex(ctx context.Context, table string, keys ...bson.E) (string, error) {
	coll := mc.db.Collection(table)
	indexKeys := make(bson.D, 0)
	for _, key := range keys {
		indexKeys = append(indexKeys, key)
	}

	model := mongo.IndexModel{Keys: indexKeys}

	index, err := coll.Indexes().CreateOne(ctx, model)

	return index, err
}

func (mc *mongoClient) DeleteOne(ctx context.Context, table string, filter interface{}, opts ...*options.DeleteOptions) error {
	coll := mc.db.Collection(table)

	_, err := coll.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (mc *mongoClient) DeleteMany(ctx context.Context, table string, filter interface{}, opts ...*options.DeleteOptions) error {
	coll := mc.db.Collection(table)

	_, err := coll.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return err
	}

	return nil
}
