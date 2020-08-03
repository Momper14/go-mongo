package main

import (
	"fmt"
	"go-mongodriver/mongoclient"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
)

type Test struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string
}

func main() {
	config := mongoclient.NewClientConfig()

	config.Database = "test"

	// make a new Object
	client, err := mongoclient.New(config)

	if err != nil {
		fmt.Println(err)
	}

	e := Test{Name: "foo"}

	//add it to db
	fmt.Println("Insert: ")
	if err = client.Insert(e, &e); err != nil {
		log.Fatal(err)
	}

	fmt.Println(e)

	//update it
	fmt.Println("Save: ")
	e.Name = "bar"
	if err = client.Save(e, &e); err != nil {
		log.Fatal(err)
	}

	fmt.Println(e)

	//find All
	fmt.Println("Find All: ")
	var entities []Test
	if err := client.FindAll(&entities); err != nil {
		log.Fatal(err)
	}
	for _, result := range entities {
		fmt.Printf("\t%v\n", result)
	}

	//delete it
	fmt.Println("Delete: ")
	if err = client.Delete(e); err != nil {
		log.Fatal(err)
	}
	fmt.Println("successfully deleted")

	fmt.Println("Try to find the deleted Document: ")
	if err = client.FindByID(e.ID, &e); err != nil {
		fmt.Println(err)
	} else {
		log.Fatal(e)
	}
}

