package task

import (
	"time"
)

//go:generate dynamodb-repo Task
//go:generate gofmt -w ./

// Task 拡張インデックスなし
type Task struct {
	ID         int64     `dynamo:"id,hash"`
	Desc       string    `dynamo:"description"`
	Created    time.Time `dynamo:"created"`
	Done       bool      `dynamo:"done"`
	Done2      bool      `dynamo:"done2"`
	Count      int       `dynamo:"count"`
	Count64    int64     `dynamo:"count64"`
	NameList   []string  `dynamo:"nameList"`
	Proportion float64   `dynamo:"proportion"`
	Flag       Flag      `dynamo:"flag"` // NG
}

type Flag bool
