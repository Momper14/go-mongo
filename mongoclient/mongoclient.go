package mongoclient

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ClientConfig config for Client
type ClientConfig struct {
	Database string
	Host     string
	Port     string
}

// Client for MongoDB
type Client struct {
	database    *mongo.Database
	collections map[string]*mongo.Collection
	client      *mongo.Client
}

// NewClientConfig creates a new ClientConfig with default values for host and port
func NewClientConfig() ClientConfig {
	return ClientConfig{
		Host: "127.0.0.1",
		Port: "27017",
	}
}

func (c ClientConfig) url() string {
	return fmt.Sprintf("mongodb://%s:%s", c.Host, c.Port)
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

func (c Client) findByID(id interface{}, ptrResult interface{}) error {
	if !isPointerOfStruct(ptrResult) {
		return fmt.Errorf("result must be a pointer")
	}

	return c.findByIDFrom(id, ptrResult, reflect.TypeOf(ptrResult).Elem().Name())
}

func (c Client) findByIDFrom(id interface{}, ptrResult interface{}, collectionName string) error {

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

	return c.findByIDFrom(insertResult.InsertedID, ptrResult, collectionName)
}

func (c Client) saveTo(entity interface{}, ptrResult interface{}, collectionName string) error {

	var (
		id     reflect.Value
		exists bool
		err    error
		ok     bool
	)
	if !isStruct(entity) {
		return fmt.Errorf("entity must be a struct")
	}
	if !isPointerOfStruct(ptrResult) {
		return fmt.Errorf("result must be a pointer")
	}

	if id, ok = structFieldValueByTag(entity, "bson", "_id"); !ok {
		//TODO: add support for structs without _id
		return fmt.Errorf("the struct must have an _id Field")
	}
	//if the object id is nil, we consider the object as new object and insert it into db
	if id.Interface() == nil {
		return c.InsertInto(entity, ptrResult, collectionName)
	}
	//if there exist an id,we check if the object is given in db, if so we update it, otherwise we insert it as a new object
	if exists, err = c.ExistsIn(id, collectionName); err != nil {
		return err
	}
	if !exists {
		return c.InsertInto(entity, ptrResult, collectionName)
	}
	// TODO: do update the object
}

func arrayContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func structFieldValueByTag(s interface{}, tagKey, tagValue string) (reflect.Value, bool) {
	rt := reflect.TypeOf(s)
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		v := field.Tag.Get(tagKey)
		arr := strings.Split(v, ",")
		if arrayContains(arr, tagValue) {
			return reflect.ValueOf(s).FieldByIndex(field.Index), true
		}
	}
	return reflect.Value{}, false
}

func isStruct(i interface{}) bool {
	if reflect.TypeOf(i).Kind() == reflect.Struct {
		return true
	}

	return false
}

func isPointerOfStruct(i interface{}) bool {
	typeOf := reflect.TypeOf(i)

	if typeOf.Kind() == reflect.Ptr {
		return typeOf.Elem().Kind() == reflect.Struct
	}

	return false
}

func (c Client) ExistsIn(id interface{}, collectionName string) (bool, error) {

	collection := c.getCollection(collectionName)
	opt := options.Find()
	opt.SetLimit(1)
	if err := collection.FindOne(context.Background(), bson.M{"_id": id}).Err(); err != nil {
		if err != mongo.ErrNoDocuments {
			return false, err
		}
		return false, nil
	}
	return true, nil
}
