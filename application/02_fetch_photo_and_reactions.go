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
	TABLE     = "quick-photos"
	USER      = "david25"
	TIMESTAMP = "2019-03-02T09:11:30"
)

type Photo struct {
	Username  string
	Timestamp string
	Location  string
	Reactions []Reaction
}

func (p *Photo) String() string {
	return fmt.Sprintf("Photo<%s -- %s>", p.Username, p.Timestamp)
}

type Reaction struct {
	ReactingUser string
	Photo        string
	ReactionType string
	Timestamp    string
}

func (r *Reaction) String() string {
	return fmt.Sprintf("Reaction<%s -- %s -- %s>", r.ReactingUser, r.Photo, r.ReactionType)
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

func NewPhotoFromDynamoDbQueryResult(out *dynamodb.QueryOutput) *Photo {
	if len(out.Items) == 0 {
		return &Photo{}
	}
	// reverse items
	items := out.Items
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	photoItem := items[0]
	photo := Photo{
		Username:  aws.StringValue(photoItem["username"].S),
		Timestamp: aws.StringValue(photoItem["timestamp"].S),
		Location:  aws.StringValue(photoItem["location"].S),
	}

	reactionItems := items[1:]

	reactions := make([]Reaction, 0)
	for _, item := range reactionItems {
		reactions = append(
			reactions,
			Reaction{
				ReactingUser: aws.StringValue(item["reactingUser"].S),
				Photo:        aws.StringValue(item["photo"].S),
				ReactionType: aws.StringValue(item["reactionType"].S),
				Timestamp:    aws.StringValue(item["timestamp"].S),
			},
		)
	}
	photo.Reactions = reactions

	return &photo
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
			"SK = :sk AND PK BETWEEN :reactions AND :user",
		),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(fmt.Sprintf("PHOTO#%s#%s", USER, TIMESTAMP)),
			},
			":reactions": {
				S: aws.String("REACTION#"),
			},
			":user": {
				S: aws.String("USER$"),
			},
		},
		ScanIndexForward: aws.Bool(true),
	}

	resp, err := db.Client().Query(&query)
	if err != nil {
		fmt.Print("Index is still backfilling. Please try again in a mount")
		panic(err)
	}
	photo := NewPhotoFromDynamoDbQueryResult(resp)
	fmt.Println(photo)
	for _, r := range photo.Reactions {
		fmt.Println(&r)
	}

	// "github.com/guregu/dynamo" を使った場合は map ではなく、struct として取得できる
	quickPhotos := make([]QuickPhoto, 0)
	t.Get("SK", fmt.Sprintf("PHOTO#%s#%s", USER, TIMESTAMP)).
		Range("PK", dynamo.Between, "REACTION#", "USER$").
		Index("InvertedIndex").
		All(&quickPhotos)
	fmt.Println(quickPhotos)
}
