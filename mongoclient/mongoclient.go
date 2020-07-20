package mongoclient

import (
	"context"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ClientConfig config for Client
type ClientConfig struct {
	Database string
	Host     string
	Port     string
}

// Client client for MongoDB
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

// Insert insert the entity to a collection of it's struct name and returns the id
// entity: entity to insert
func (c Client) Insert(entity interface{}) (interface{}, error) {
	return c.InsertInto(entity, reflect.TypeOf(entity).Name())
}

// InsertInto insert the entity into the given collection and returns the id
// entity: entity to insert
// name: name of the collection
func (c Client) InsertInto(entity interface{}, name string) (interface{}, error) {

	typeOf := reflect.TypeOf(entity)

	if typeOf.Kind() == reflect.Ptr {
		typeOf = reflect.TypeOf(&entity)
	}

	if typeOf.Kind() != reflect.Struct {
		return "", fmt.Errorf("entity must be a struct")
	}

	collection := c.getCollection(name)

	insertResult, err := collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return "", err
	}

	return insertResult.InsertedID, nil
}
