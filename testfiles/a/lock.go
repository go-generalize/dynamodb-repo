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
	MetaPayload Meta
}
type Nest1Type struct {
	Nest1 Nest2Type
}
type Lock struct {
	ID    int64  `dynamo:"id,hash"     validate:"required"`
	Name  string `dynamo:"name,unique"`
	Name2 string `dynamo:"name2,unique"`
	Email string `dynamo:"email"       validate:"email"`
	Meta  Nest1Type
}
