package task

const NameSchema = `
{
	"AttributeDefinitions": [
		{
			"AttributeName": "id",
			"AttributeType": "N"
		},
		{
			"AttributeName": "description",
			"AttributeType": "S"
		},
		{
			"AttributeName": "description2",
			"AttributeType": "S"
		},
		{
			"AttributeName": "created",
			"AttributeType": "S"
		},
		{
			"AttributeName": "done",
			"AttributeType": "BOOL"
		},
		{
			"AttributeName": "count",
			"AttributeType": "N"
		},
		{
			"AttributeName": "priceList",
			"AttributeType": "NS"
		}
	],
	"TableName": "Name",
	"KeySchema": [
		{
			"AttributeName": "id",
			"KeyType": "HASH"
		}
	],
	"GlobalSecondaryIndexes": [
		{
			"IndexName": "count-index",
			"ProvisionedThroughput": {
				"WriteCapacityUnits": 5,
				"ReadCapacityUnits": 5
			},
			"KeySchema": [
				{
					"KeyType": "HASH",
					"AttributeName": "count"
				}
			]
		},
		{
			"IndexName": "description-index",
			"ProvisionedThroughput": {
				"WriteCapacityUnits": 5,
				"ReadCapacityUnits": 5
			},
			"KeySchema": [
				{
					"KeyType": "HASH",
					"AttributeName": "description"
				}
			]
		},
		{
			"IndexName": "description2-index",
			"ProvisionedThroughput": {
				"WriteCapacityUnits": 5,
				"ReadCapacityUnits": 5
			},
			"KeySchema": [
				{
					"KeyType": "HASH",
					"AttributeName": "description2"
				}
			]
		},
		{
			"IndexName": "done-index",
			"ProvisionedThroughput": {
				"WriteCapacityUnits": 5,
				"ReadCapacityUnits": 5
			},
			"KeySchema": [
				{
					"KeyType": "HASH",
					"AttributeName": "done"
				}
			]
		},
		{
			"IndexName": "created-index",
			"ProvisionedThroughput": {
				"WriteCapacityUnits": 5,
				"ReadCapacityUnits": 5
			},
			"KeySchema": [
				{
					"KeyType": "HASH",
					"AttributeName": "created"
				}
			]
		}
	],
	"ProvisionedThroughput": {
		"ReadCapacityUnits": 5,
		"WriteCapacityUnits": 5
	}
}
`

const TaskSchema = `
{
	"AttributeDefinitions": [
		{
			"AttributeName": "id",
			"AttributeType": "N"
		},
		{
			"AttributeName": "description",
			"AttributeType": "S"
		},
		{
			"AttributeName": "created",
			"AttributeType": "S"
		},
		{
			"AttributeName": "done",
			"AttributeType": "BOOL"
		},
		{
			"AttributeName": "done2",
			"AttributeType": "BOOL"
		},
		{
			"AttributeName": "count",
			"AttributeType": "N"
		},
		{
			"AttributeName": "count64",
			"AttributeType": "N"
		},
		{
			"AttributeName": "nameList",
			"AttributeType": "SS"
		},
		{
			"AttributeName": "proportion",
			"AttributeType": "N"
		},
		{
			"AttributeName": "flag",
			"AttributeType": "BOOL"
		}
	],
	"TableName": "Task",
	"KeySchema": [
		{
			"AttributeName": "id",
			"KeyType": "HASH"
		}
	],
	"GlobalSecondaryIndexes": [
		{
			"IndexName": "count-index",
			"ProvisionedThroughput": {
				"WriteCapacityUnits": 5,
				"ReadCapacityUnits": 5
			},
			"KeySchema": [
				{
					"KeyType": "HASH",
					"AttributeName": "count"
				}
			]
		},
		{
			"IndexName": "proportion-index",
			"ProvisionedThroughput": {
				"WriteCapacityUnits": 5,
				"ReadCapacityUnits": 5
			},
			"KeySchema": [
				{
					"KeyType": "HASH",
					"AttributeName": "proportion"
				}
			]
		},
		{
			"IndexName": "done-index",
			"ProvisionedThroughput": {
				"WriteCapacityUnits": 5,
				"ReadCapacityUnits": 5
			},
			"KeySchema": [
				{
					"KeyType": "HASH",
					"AttributeName": "done"
				}
			]
		},
		{
			"IndexName": "created-index",
			"ProvisionedThroughput": {
				"WriteCapacityUnits": 5,
				"ReadCapacityUnits": 5
			},
			"KeySchema": [
				{
					"KeyType": "HASH",
					"AttributeName": "created"
				}
			]
		}
	],
	"ProvisionedThroughput": {
		"ReadCapacityUnits": 5,
		"WriteCapacityUnits": 5
	}
}
`
