//go:build ignore

// https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/4/

package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

type QuickPhoto struct {
	PK string `dynamo:"PK,hash"`
	SK string `dynamo:",range"`
}

func main() {
	sess := session.Must(session.NewSession())
	db := dynamo.New(
		sess,
		&aws.Config{
			Region: aws.String("ap-northeast-1"),
			// LogLevel: aws.LogLevel(aws.LogDebug),
		},
	)
	table := QuickPhoto{}

	ct := db.CreateTable(
		"quick-photos", // tableName
		table,
	)

	ct.OnDemand(false)
	ct.Provision(10, 10)
	err := ct.Run()
	if err != nil {
		fmt.Print("Could not create table. Error: ")
		panic(err)
	}

	fmt.Println("Table created successfully.")

}
