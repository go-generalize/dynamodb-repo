package model

import (
	"time"
)

//go:generate dynamodb-repo Name

type Name struct {
	ID       string    `dynamo:",hash"`
	Desc     string    `dynamo:"description,range"`
	Created  time.Time `dynamo:"created"`
	Done     bool      `dynamo:"done"`
	Count    int       `dynamo:"count"`
	NameList []string  `dynamo:"nameList"`
	// TODO Indexes  []string  `dynamo:"indexes"`
}
