package main

import (
	"fmt"
	"go-mongodriver/mongoclient"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Test struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string
}

func main() {
	config := mongoclient.NewClientConfig()

	config.Database = "test"

	client, err := mongoclient.New(config)

	if err != nil {
		fmt.Println(err)
	}

	e := Test{Name: "foo"}

	if err = client.Insert(e, &e); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(e)
	}

}
