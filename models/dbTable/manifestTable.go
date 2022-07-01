package dbTable

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pennsieve/pennsieve-go-api/models/manifest"
)

// ManifestTable is a representation of a Manifest in DynamoDB
type ManifestTable struct {
	ManifestId     string `dynamodbav:"ManifestId"`
	DatasetId      int64  `dynamodbav:"DatasetId"`
	DatasetNodeId  string `dynamodbav:"DatasetNodeId"`
	OrganizationId int64  `dynamodbav:"OrganizationId"`
	UserId         int64  `dynamodbav:"UserId"`
	Status         string `dynamodbav:"Status"`
}

// ManifestFileTable is a representation of a ManifestFile in DynamoDB
type ManifestFileTable struct {
	ManifestId string `dynamodbav:"ManifestId"`
	UploadId   string `dynamodbav:"UploadId"`
	FilePath   string `dynamodbav:"FilePath,omitempty"`
	FileName   string `dynamodbav:"FileName"`
	Status     string `dynamodbav:"Status"`
}

type ManifestFilePrimaryKey struct {
	ManifestId string `dynamodbav:"ManifestId"`
	UploadId   string `dynamodbav:"UploadId"`
}

// GetFromManifest returns a Manifest item for a given manifest ID.
func GetFromManifest(client *dynamodb.Client, manifestTableName string, manifestId string) (*ManifestTable, error) {

	item := ManifestTable{}

	data, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(manifestTableName),
		Key: map[string]types.AttributeValue{
			"ManifestId": &types.AttributeValueMemberS{Value: manifestId},
		},
	})

	if err != nil {
		return &item, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return &item, fmt.Errorf("GetItem: Manifest not found.\n")
	}

	err = attributevalue.UnmarshalMap(data.Item, &item)
	if err != nil {
		return &item, fmt.Errorf("UnmarshalMap: %v\n", err)
	}

	return &item, nil

}

// UpdateFileTableStatus updates the status of the file in the file-table dynamodb
func UpdateFileTableStatus(client *dynamodb.Client, tableName string, manifestId string, uploadId string, status manifest.ManifestFileStatus) error {

	_, err := client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ManifestId": &types.AttributeValueMemberS{Value: manifestId},
			"UploadId":   &types.AttributeValueMemberS{Value: uploadId},
		},
		UpdateExpression: aws.String("set #status = :statusValue"),
		ExpressionAttributeNames: map[string]string{
			"#status": "Status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":statusValue": &types.AttributeValueMemberS{Value: status.String()},
		},
	})
	return err
}
