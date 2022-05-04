//go:build ignore

// https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/5/

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
	USER  = "haroldwatkins"
)

type Friendship struct {
	FollowedUser  string
	FollowingUser string
	Timestamp     string
}

func (r Friendship) String() string {
	return fmt.Sprintf("Friendship<%s -- %s>", r.FollowedUser, r.FollowingUser)
}

type QuickPhoto struct {
	PK            string   `dynamo:"PK,hash" json:"PK"`
	SK            string   `dynamo:",range" json:"SK"`
	Address       string   `dynamo:"address" json:"address"`
	Birthdate     string   `dynamo:"birthdate" json:"birthdate"`
	Email         string   `dynamo:"email" json:"email"`
	Name          string   `dynamo:"name" json:"name"`
	Username      string   `dynamo:"username" json:"username"`
	Status        string   `dynamo:"status" json:"status"`
	Interests     []string `dynamo:"interests" json:"interests"`
	Followers     int      `dynamo:"followers" json:"followers"`
	Following     int      `dynamo:"following" json:"following"`
	PinnedImage   string   `dynamo:"pinnedImage" json:"pinnedImage"`
	Timestamp     string   `dynamo:"timestamp" json:"timestamp"`
	FollowedUser  string   `dynamo:"followedUser" json:"followedUser"`
	FollowingUser string   `dynamo:"followingUser" json:"followingUser"`
	Location      string   `dynamo:"location" json:"location"`
	Reactions     struct {
		PlusOne    int `dynamo:"+1" json:"+1"`
		Smiley     int `dynamo:"smiley" json:"smiley"`
		Sunglasses int `dynamo:"sunglasses" json:"sunglasses"`
		Heart      int `dynamo:"heart" json:"heart"`
	} `dynamo:"reactions" json:"reactions"`
	ReactingUser string `dynamo:"reactingUser" json:"reactingUser"`
	Photo        string `dynamo:"photo" json:"photo"`
	ReactionType string `dynamo:"reactionType" json:"reactionType"`
}

func NewFriendshipsFromDynamoDbQueryResult(out *dynamodb.QueryOutput) []Friendship {
	friendships := make([]Friendship, 0)
	if len(out.Items) == 0 {
		return friendships
	}
	items := out.Items

	for _, item := range items {
		friendships = append(
			friendships,
			Friendship{
				FollowedUser:  aws.StringValue(item["followedUser"].S),
				FollowingUser: aws.StringValue(item["followingUser"].S),
				Timestamp:     aws.StringValue(item["timestamp"].S),
			},
		)
	}

	return friendships
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

	// https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/5/
	query := dynamodb.QueryInput{
		TableName: aws.String(TABLE),
		IndexName: aws.String("InvertedIndex"),
		KeyConditionExpression: aws.String(
			"SK = :sk",
		),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(fmt.Sprintf("#FRIEND#%s", USER)),
			},
		},
		ScanIndexForward: aws.Bool(true),
	}

	resp, err := db.Client().Query(&query)
	if err != nil {
		fmt.Print("Index is still backfilling. Please try again in a mount")
		panic(err)
	}
	friendships := NewFriendshipsFromDynamoDbQueryResult(resp)
	for _, friendship := range friendships {
		fmt.Println(friendship)
	}

	// "github.com/guregu/dynamo" を使った場合は map ではなく、struct として取得できる
	quickPhotos := make([]QuickPhoto, 0)
	t.Get("SK", fmt.Sprintf("#FRIEND#%s", USER)).
		Index("InvertedIndex").
		All(&quickPhotos)
	fmt.Println(quickPhotos)
}
