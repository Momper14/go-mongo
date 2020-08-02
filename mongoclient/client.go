package mongoclient

import (
	"context"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client for MongoDB
type Client struct {
	database    *mongo.Database
	collections map[string]*mongo.Collection
	client      *mongo.Client
}

// New creates a new client
func New(config ClientConfig) (Client, error) {
	var err error

	client := Client{collections: make(map[string]*mongo.Collection)}
	clientOptions := options.Client().ApplyURI(config.url())

	// Connect to MongoDB
	if client.client, err = mongo.Connect(context.TODO(), clientOptions); err != nil {
		return client, err
	}

	// Check the connection
	if err = client.client.Ping(context.TODO(), nil); err != nil {
		return client, err
	}

	client.database = client.client.Database(config.Database)

	return client, nil
}

func (c Client) getCollection(name string) *mongo.Collection {
	if collection, ok := c.collections[name]; ok {
		return collection
	}

	collection := c.database.Collection(name)
	c.collections[name] = collection

	return collection
}

func (c Client) FindByID(id interface{}, ptrResult interface{}) error {
	if !isPointerOfStruct(ptrResult) {
		return fmt.Errorf("result must be a pointer")
	}

	return c.FindByIDFrom(id, ptrResult, reflect.TypeOf(ptrResult).Elem().Name())
}

func (c Client) FindByIDFrom(id interface{}, ptrResult interface{}, collectionName string) error {

	if !isPointerOfStruct(ptrResult) {
		return fmt.Errorf("result must be a pointer")
	}

	collection := c.getCollection(collectionName)

	return collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(ptrResult)
}

// Insert insert the entity to a collection of it's struct name and returns the id
// entity: entity to insert
func (c Client) Insert(entity interface{}, ptrResult interface{}) error {
	return c.InsertInto(entity, ptrResult, reflect.TypeOf(entity).Name())
}

// InsertInto insert the entity into the given collection and returns the id
// entity: entity to insert
// name: name of the collection
func (c Client) InsertInto(entity interface{}, ptrResult interface{}, collectionName string) error {

	if !isStruct(entity) {
		return fmt.Errorf("entity must be a struct")
	}

	if !isPointerOfStruct(ptrResult) {
		return fmt.Errorf("result must be a pointer")
	}

	collection := c.getCollection(collectionName)

	insertResult, err := collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return err
	}

	return c.FindByIDFrom(insertResult.InsertedID, ptrResult, collectionName)
}

func (c Client) Save(entity interface{}, ptrResult interface{}) error {
	return c.SaveTo(entity, ptrResult, reflect.TypeOf(entity).Name())
}

func (c Client) SaveTo(entity interface{}, ptrResult interface{}, collectionName string) error {

	var (
		id     interface{}
		exists bool
		err    error
	)

	if !isStruct(entity) {
		return fmt.Errorf("entity must be a struct")
	}

	if !isPointerOfStruct(ptrResult) {
		return fmt.Errorf("result must be a pointer")
	}

	if idTmp, ok := structFieldValueByTag(entity, "bson", "_id"); ok {
		id = idTmp.Interface()
	} else {
		//TODO: add support for structs without _id
		return fmt.Errorf("the struct must have an _id Field")
	}

	//if the object id is nil, we consider the object as new object and insert it into db
	if id == nil {
		return c.InsertInto(entity, ptrResult, collectionName)
	}

	//if there exist an id,we check if the object is given in db, if so we update it, otherwise we insert it as a new object
	if exists, err = c.ExistsIn(entity, collectionName); err != nil {
		return err
	}

	if !exists {
		return c.InsertInto(entity, ptrResult, collectionName)
	}

	collection := c.getCollection(collectionName)
	if result, err := collection.ReplaceOne(context.Background(), bson.M{"_id": id}, entity); err != nil {
		return err
	} else if result.MatchedCount == 0 {
		return fmt.Errorf("no document modified")
	}

	return c.FindByIDFrom(id, ptrResult, collectionName)
}

func (c Client) Exists(entity interface{}) (bool, error) {
	return c.ExistsIn(entity, reflect.TypeOf(entity).Name())
}

func (c Client) ExistsIn(entity interface{}, collectionName string) (bool, error) {

	var (
		id    interface{}
		count int64
		err   error
	)

	if !isStruct(entity) {
		return false, fmt.Errorf("entity must be a struct")
	}

	if idTmp, ok := structFieldValueByTag(entity, "bson", "_id"); ok {
		id = idTmp.Interface()
	} else {
		//TODO: add support for structs without _id
		return false, fmt.Errorf("the struct must have an _id Field")
	}

	collection := c.getCollection(collectionName)

	if count, err = collection.CountDocuments(context.Background(), bson.M{"_id": id}); err != nil {
		return false, err
	}

	return count >= 1, nil
}
