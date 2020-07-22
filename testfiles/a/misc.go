// THIS FILE IS A GENERATED CODE. DO NOT EDIT
// generated version: 0.3.0
package model

type GetOption struct {
	IncludeSoftDeleted bool
}

type DeleteMode string

const (
	DeleteModeSoft = "soft"
	DeleteModeHard = "hard"
)

type DeleteOption struct {
	Mode DeleteMode
}