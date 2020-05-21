package task

import (
	"time"
)

//go:generate dynamodb-repo Name
//go:generate gofmt -w ./

// Name 拡張インデックスあり
type Name struct {
	ID        int64     `dynamo:"-,hash"`
	Created   time.Time `dynamo:"created"`
	Desc      string    `dynamo:"description"`
	Desc2     string    `dynamo:"description2"`
	Done      bool      `dynamo:"done"`
	Count     int       `dynamo:"count"`
	PriceList []int     `dynamo:"priceList"`
}
