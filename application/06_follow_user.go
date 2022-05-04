//go:build ignore

// https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/7/

package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
)

const (
	TABLE          = "quick-photos"
	FOLLOWED_USER  = "tmartinez"
	FOLLOWING_USER = "john42"
)

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

func main() {
	sess := session.Must(session.NewSession())
	db := dynamo.New(
		sess,
		&aws.Config{
			Region: aws.String("ap-northeast-1"),
		},
	)
	t := db.Table(TABLE)

	userStr := fmt.Sprintf("USER#%s", FOLLOWED_USER)
	frindStr := fmt.Sprintf("#FRIEND#%s", FOLLOWING_USER)
	userMetadataStr := fmt.Sprintf("#METADATA#%s", FOLLOWED_USER)
	friendUserStr := fmt.Sprintf("USER#%s", FOLLOWING_USER)
	friendMetadataStr := fmt.Sprintf("#METADATA#%s", FOLLOWING_USER)
	now := time.Now().Format("2006-01-02T15:04:05")

	items := []*dynamodb.TransactWriteItem{
		{
			Put: &dynamodb.Put{
				TableName: aws.String(TABLE),
				Item: map[string]*dynamodb.AttributeValue{
					"PK": {
						S: aws.String(userStr),
					},
					"SK": {
						S: aws.String(frindStr),
					},
					"followedUser": {
						S: aws.String(FOLLOWED_USER),
					},
					"followingUser": {
						S: aws.String(FOLLOWING_USER),
					},
					"timestamp": {
						S: aws.String(now),
					},
				},
				ConditionExpression:                 aws.String("attribute_not_exists(SK)"),
				ReturnValuesOnConditionCheckFailure: aws.String(dynamodb.ReturnValueAllOld),
			},
		},
		{
			Update: &dynamodb.Update{
				TableName: aws.String(TABLE),
				Key: map[string]*dynamodb.AttributeValue{
					"PK": {
						S: aws.String(userStr),
					},
					"SK": {
						S: aws.String(userMetadataStr),
					},
				},
				UpdateExpression: aws.String(
					"SET followers = followers + :i",
				),
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":i": {
						N: aws.String("1"),
					},
				},
				ReturnValuesOnConditionCheckFailure: aws.String(dynamodb.ReturnValueAllOld),
			},
		},
		{
			Update: &dynamodb.Update{
				TableName: aws.String(TABLE),
				Key: map[string]*dynamodb.AttributeValue{
					"PK": {
						S: aws.String(friendUserStr),
					},
					"SK": {
						S: aws.String(friendMetadataStr),
					},
				},
				UpdateExpression: aws.String(
					"SET following = following + :i",
				),
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":i": {
						N: aws.String("1"),
					},
				},
				ReturnValuesOnConditionCheckFailure: aws.String(dynamodb.ReturnValueAllOld),
			},
		},
	}
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	}
	c := db.Client()
	_, err := c.TransactWriteItems(input)
	if err != nil {
		fmt.Print("Could not add follow relationship Err:")
		panic(err)
	}
	fmt.Println(fmt.Sprintf("User %s is now following user %s", FOLLOWING_USER, FOLLOWED_USER))

	// "github.com/guregu/dynamo" を使った場合
	// "SET reactions.#t = reactions.#t + :i" の実現に
	now2 := time.Now().Format("2006-01-02T15:04:05+09:00")
	tx := db.WriteTx()
	put := t.Put(
		QuickPhoto{
			PK:            userStr,
			SK:            friendUserStr,
			FollowedUser:  FOLLOWED_USER,
			FollowingUser: FOLLOWING_USER,
			Timestamp:     now2,
		},
	)
	update1 := t.Update("PK", userStr).
		Range("SK", userMetadataStr).
		SetExpr("followed = followed + ?", 1)
	update2 := t.Update("PK", friendUserStr).
		Range("SK", friendMetadataStr).
		SetExpr("followed = followed + ?", 1)
	err = tx.
		Put(put).
		Update(update1).
		Update(update2).
		Run()
	if err != nil {
		panic(err)
	}
}
