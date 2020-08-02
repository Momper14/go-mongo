package main

import (
	"fmt"
	"go-mongodriver/mongoclient"
	"log"

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
		log.Fatal(err)
	}

	fmt.Println(e)

	e.Name = "bar"

	if err = client.Save(e, &e); err != nil {
		log.Fatal(err)
	}

	fmt.Println(e)

}
