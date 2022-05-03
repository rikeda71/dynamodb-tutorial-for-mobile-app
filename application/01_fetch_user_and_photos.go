//go:build ignore

// https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/4/

package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
)

const (
	TABLE = "quick-photos"
	USER  = "jacksonjason"
)

type User struct {
	Username           string
	Name               string
	Email              string
	Birthdate          string
	Address            string
	Status             string
	Interests          string
	PinnedImage        string
	RecommendedFriends []string
	Photos             []Photo
}

type Photo struct {
	Username  string
	Timestamp string
	Location  string
}

func NewUserFromDynamoDbQueryResult(out *dynamodb.QueryOutput) *User {
	if len(out.Items) == 0 {
		return &User{}
	}
	userItem := out.Items[0]

	fmt.Println(out.Items[0])
	user := User{
		Username:  aws.StringValue(userItem["username"].S),
		Name:      aws.StringValue(userItem["name"].S),
		Email:     aws.StringValue(userItem["email"].S),
		Birthdate: aws.StringValue(userItem["birthdate"].S),
		Address:   aws.StringValue(userItem["address"].S),
		Status:    aws.StringValue(userItem["status"].S),
		Interests: aws.StringValue(userItem["interests"].S),
	}

	if val, ok := userItem["pinnedImage"]; ok {
		user.PinnedImage = aws.StringValue(val.S)
	}

	if val, ok := userItem["reccomendedFriends"]; ok {
		friends := make([]string, 0)
		for _, v := range val.L {
			friends = append(friends, aws.StringValue(v.S))
		}
		user.RecommendedFriends = friends
	}

	photoItems := out.Items[1:]

	photos := make([]Photo, 0)
	for _, item := range photoItems {
		photos = append(
			photos,
			Photo{
				Username:  aws.StringValue(item["username"].S),
				Timestamp: aws.StringValue(item["timestamp"].S),
				Location:  aws.StringValue(item["location"].S),
			},
		)
	}
	user.Photos = photos

	return &user
}

func main() {
	sess := session.Must(session.NewSession())
	db := dynamo.New(
		sess,
		&aws.Config{
			Region: aws.String("ap-northeast-1"),
		},
	)
	t := db.Table(TABLE)

	// https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/4/
	query := dynamodb.QueryInput{
		TableName: aws.String(TABLE),
		KeyConditionExpression: aws.String(
			"PK = :pk AND SK BETWEEN :metadata AND :photos",
		),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(fmt.Sprintf("USER#%s", USER)),
			},
			":metadata": {
				S: aws.String(fmt.Sprintf("#METADATA#%s", USER)),
			},
			":photos": {
				S: aws.String("PHOTO$"),
			},
		},
		ScanIndexForward: aws.Bool(true),
	}

	resp, err := db.Client().Query(&query)
	if err != nil {
		panic(err)
	}
	user := NewUserFromDynamoDbQueryResult(resp)
	fmt.Println(user)
	// TODO
	// var v interface{}
	// t.Get("PK", fmt.Sprintf("USER#%s", USER)).
	// 	Range("SK", dynamo.Between, fmt.Sprintf("#METADATA%s", USER), aws.String("PHOTO$")).
	// 	All(&v)
	// fmt.Println(v)
}
