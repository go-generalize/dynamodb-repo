package model

import (
	"time"
)

//go:generate dynamodb-repo -prefix=Prefix Lock

type Meta struct {
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt time.Time
	UpdatedBy string
	DeletedAt *time.Time
	DeletedBy string
	Version   int
}

// Lock Metaテスト用
type Nest2Type struct {
	Meta
}
type Nest1Type struct {
	Nest2Type
}
type Lock struct {
	ID   int64 `dynamo:"id,hash"`
	Name string
	Nest1Type
}
