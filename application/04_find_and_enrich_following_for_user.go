//go:build ignore

// https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/6/

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
}

func (u User) String() string {
	return fmt.Sprintf("User<%s -- %s>", u.Username, u.Name)
}

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

func NewUsersFromDynamoDbAttributeValues(avs []map[string]*dynamodb.AttributeValue) []User {
	users := make([]User, 0)
	for _, userItem := range avs {
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

		users = append(users, user)
	}

	return users
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

	api := db.Client()
	resp, err := api.Query(&query)
	if err != nil {
		fmt.Print("Index is still backfilling. Please try again in a mount")
		panic(err)
	}
	friendships := NewFriendshipsFromDynamoDbQueryResult(resp)

	// 部分正規化のための処理
	// BatchGetItem を使って、該当するユーザー情報を並列に1つずつ取得する
	keys := make([]map[string]*dynamodb.AttributeValue, 0)
	for _, friendship := range friendships {
		keys = append(keys, map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(fmt.Sprintf("USER#%s", friendship.FollowedUser)),
			},
			"SK": {
				S: aws.String(fmt.Sprintf("#METADATA#%s", friendship.FollowedUser)),
			},
		})
	}
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			TABLE: {
				Keys: keys,
			},
		},
	}
	batchResults, err := api.BatchGetItem(input)
	if err != nil {
		panic(err)
	}
	users := NewUsersFromDynamoDbAttributeValues(batchResults.Responses[TABLE])
	for _, user := range users {
		fmt.Println(user)
	}

	// "github.com/guregu/dynamo" を使った場合は map ではなく、struct として取得できる
	quickPhotos := make([]QuickPhoto, 0)
	t.Get("SK", fmt.Sprintf("#FRIEND#%s", USER)).
		Index("InvertedIndex").
		All(&quickPhotos)

	batchedQuickPhotos := make([]QuickPhoto, 0)
	dynamoKeys := make([]dynamo.Keyed, 0)
	for _, qp := range quickPhotos {
		dynamoKeys = append(dynamoKeys, dynamo.Keys{fmt.Sprintf("USER#%s", qp.FollowedUser), fmt.Sprintf("#METADATA#%s", qp.FollowedUser)})
	}
	t.Batch("PK", "SK").Get(dynamoKeys...).All(&batchedQuickPhotos)

	for _, bqp := range batchedQuickPhotos {
		fmt.Println(bqp)
	}
}
