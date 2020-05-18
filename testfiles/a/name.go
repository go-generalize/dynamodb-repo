package task

import (
	"time"
)

//go:generate dynamodb-repo Name
//go:generate gofmt -w ./

// Name 拡張インデックスあり
type Name struct {
	ID      int64     `dynamo:"-,hash"`
	Created time.Time `dynamo:"created"`
	// supported indexer tags word: e/equal(Default), l/like, p/prefix,
	// TODO s/suffix
	Desc      string   `dynamo:"description" indexer:"l"`
	Desc2     string   `dynamo:"description2" indexer:"p"`
	Done      bool     `dynamo:"done"`
	Count     int      `dynamo:"count"`
	PriceList []int    `dynamo:"priceList"`
	Indexes   []string `dynamo:"indexes"`
}
