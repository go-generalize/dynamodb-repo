package model

import dda "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

//go:generate dynamodb-repo Name

type CustomStruct struct {
	Value int
	Str   string
}

// Name RangeKeyあり
type Name struct {
	ID        int64           `dynamo:"id,hash" auto:""`
	Count     int             `dynamo:"count,range"`
	Created   dda.UnixTime    `dynamo:"created"`
	Desc      string          `dynamo:"description"`
	Desc2     string          `dynamo:"description2"`
	Done      bool            `dynamo:"done"`
	PriceList []int           `dynamo:"priceList"`
	Value     CustomStruct    `dynamo:"custom"`
	Array     []*CustomStruct `dynamo:"customs"`
}
