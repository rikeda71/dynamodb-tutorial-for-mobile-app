//go:build ignore

// https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/5/

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
		},
	)

	t := db.Table("quick-photos")
	_, err := t.UpdateTable().
		CreateIndex(
			dynamo.Index{
				Name:           "InvertedIndex",
				HashKey:        "SK",
				HashKeyType:    dynamo.StringType,
				RangeKey:       "PK",
				RangeKeyType:   dynamo.StringType,
				ProjectionType: dynamo.AllProjection,
				Throughput: dynamo.Throughput{
					Read:  10,
					Write: 10,
				},
			},
		).Run()
	if err != nil {
		fmt.Print("Could not update table. Err:")
		panic(err)
	}

	fmt.Println("Table updated successfully.")

}
