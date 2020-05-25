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
			"AttributeName": "count",
			"AttributeType": "N"
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
			],
			"Projection": {
				"ProjectionType": "ALL"
			}
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
			],
			"Projection": {
				"ProjectionType": "ALL"
			}
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
			],
			"Projection": {
				"ProjectionType": "ALL"
			}
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
			],
			"Projection": {
				"ProjectionType": "ALL"
			}
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
			"AttributeName": "created",
			"AttributeType": "S"
		},
		{
			"AttributeName": "count",
			"AttributeType": "N"
		},
		{
			"AttributeName": "proportion",
			"AttributeType": "N"
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
			],
			"Projection": {
				"ProjectionType": "ALL"
			}
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
			],
			"Projection": {
				"ProjectionType": "ALL"
			}
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
			],
			"Projection": {
				"ProjectionType": "ALL"
			}
		}
	],
	"ProvisionedThroughput": {
		"ReadCapacityUnits": 5,
		"WriteCapacityUnits": 5
	}
}
`
