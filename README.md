# DynamoDB Repo

## Quick Start

Go言語の環境が整っていることを前提とします。

`dynamodb-repo` を導入します。
```bash
$ go get github.com/go-generalize/dynamodb-repo
```

`package` と `struct` を適宜定義したファイルを作ります。

<details>
<summary>クリックしてファイル例を表示</summary>

```golang
package model

import dda "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

//go:generate dynamodb-repo -disable-meta Name
//NOTE: go generateの最後の引数は生成したい struct 名を指定する

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
	Array     []*CustomStruct `dynamo:"customs"`
}
```

</details>

最後に `go generate` をすると、適宜ファイルが生成されます。
```bash
$ go generate
```

ファイルが生成されたことを確認します。

```
$ ls
constant.go  misc.go  name.go  name_gen.go
```