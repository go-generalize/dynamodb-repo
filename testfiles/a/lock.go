package model

import (
	"time"
)

//go:generate dynamodb-repo -prefix=Prefix Lock

type Meta struct {
	CreatedAt time.Time  `dynamo:"created_at"`
	CreatedBy string     `dynamo:"created_by"`
	UpdatedAt time.Time  `dynamo:"updated_at"`
	UpdatedBy string     `dynamo:"updated_by"`
	DeletedAt *time.Time `dynamo:"deleted_at"`
	DeletedBy string     `dynamo:"deleted_by"`
	Version   int        `dynamo:"version"`
}

// Lock Metaテスト用
type Nest2Type struct {
	Meta
}
type Nest1Type struct {
	Nest2Type
}
type Lock struct {
	ID   int64  `dynamo:"id,hash"`
	Name string `dynamo:"name"`
	Nest1Type
}
