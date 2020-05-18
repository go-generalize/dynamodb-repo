package model

import (
	"time"
)

//go:generate dynamodb-repo Task

type Task struct {
	ID       string    `dynamo:",hash"`
	Desc     string    `dynamo:"description"`
	Created  time.Time `dynamo:"created"`
	Done     bool      `dynamo:"done"`
	Count    int       `dynamo:"count"`
	NameList []string  `dynamo:"nameList"`
}
