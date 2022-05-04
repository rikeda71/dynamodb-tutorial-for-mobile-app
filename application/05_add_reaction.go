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
	TABLE           = "quick-photos"
	REACTING_USER   = "kennedyheather"
	REACTION_TYPE   = "sunglasses"
	PHOTO_USER      = "ppierce"
	PHOTO_TIMESTAMP = "2019-04-14T08:09:34"
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

	reactionStr := fmt.Sprintf("REACTION#%s#%s", REACTING_USER, REACTION_TYPE)
	photoStr := fmt.Sprintf("PHOTO#%s#%s", PHOTO_USER, PHOTO_TIMESTAMP)
	userStr := fmt.Sprintf("USER#%s", PHOTO_USER)
	now := time.Now().Format("2006-01-02T15:04:05")
	items := []*dynamodb.TransactWriteItem{
		{
			Put: &dynamodb.Put{
				TableName: aws.String(TABLE),
				Item: map[string]*dynamodb.AttributeValue{
					"PK": {
						S: aws.String(reactionStr),
					},
					"SK": {
						S: aws.String(photoStr),
					},
					"reactingUser": {
						S: aws.String(REACTING_USER),
					},
					"reactionType": {
						S: aws.String(REACTION_TYPE),
					},
					"photo": {
						S: aws.String(photoStr),
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
						S: aws.String(photoStr),
					},
				},
				UpdateExpression: aws.String(
					"SET reactions.#t = reactions.#t + :i",
				),
				ExpressionAttributeNames: map[string]*string{
					"#t": aws.String(REACTION_TYPE),
				},
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
		fmt.Print("Exec transaction failed. Err:")
		panic(err)
	}

	// "github.com/guregu/dynamo" を使った場合
	// "SET reactions.#t = reactions.#t + :i" の実現に
	now2 := time.Now().Format("2006-01-02T15:04:05+09:00")
	tx := db.WriteTx()
	put := t.Put(
		QuickPhoto{
			PK:           reactionStr,
			SK:           photoStr,
			ReactingUser: REACTING_USER,
			ReactionType: REACTION_TYPE,
			Photo:        photoStr,
			Timestamp:    now2,
		},
	)
	update := t.Update("PK", userStr).
		Range("SK", photoStr).
		SetExpr("reactions.$ = reactions.$ + ?", REACTION_TYPE, REACTION_TYPE, 1)
	err = tx.
		Put(put).
		Update(update).
		Run()
	if err != nil {
		panic(err)
	}
}
