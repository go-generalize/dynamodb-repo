package model

import dda "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

//go:generate dynamodb-repo -prefix=Prefix -disable-meta Task

// Task RangeKeyなし
type Task struct {
	ID         int64        `dynamo:"id,hash"`
	Desc       string       `dynamo:"description"`
	Created    dda.UnixTime `dynamo:"created"`
	Done       bool         `dynamo:"done"`
	Done2      bool         `dynamo:"done2"`
	Count      int          `dynamo:"count"`
	Count64    int64        `dynamo:"count64"`
	NameList   []string     `dynamo:"nameList"`
	Proportion float64      `dynamo:"proportion"`
	Flag       Flag         `dynamo:"flag"`
	CreatedAt  dda.UnixTime `dynamo:"createdAt"`
	UpdatedAt  dda.UnixTime `dynamo:"updatedAt"`
}

type Flag bool
