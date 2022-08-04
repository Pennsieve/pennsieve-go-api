package dbTable

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pennsieve/pennsieve-go-api/models/manifest/manifestFile"
)

// ManifestTable is a representation of a Manifest in DynamoDB
type ManifestTable struct {
	ManifestId     string `dynamodbav:"ManifestId"`
	DatasetId      int64  `dynamodbav:"DatasetId"`
	DatasetNodeId  string `dynamodbav:"DatasetNodeId"`
	OrganizationId int64  `dynamodbav:"OrganizationId"`
	UserId         int64  `dynamodbav:"UserId"`
	Status         string `dynamodbav:"Status"`
	DateCreated    int64  `dynamodbav:"DateCreated"`
}

// ManifestFileTable is a representation of a ManifestFile in DynamoDB
type ManifestFileTable struct {
	ManifestId     string `dynamodbav:"ManifestId"`
	UploadId       string `dynamodbav:"UploadId"`
	FilePath       string `dynamodbav:"FilePath,omitempty"`
	FileName       string `dynamodbav:"FileName"`
	MergePackageId string `dynamodbav:"MergePackageId,omitempty"`
	Status         string `dynamodbav:"Status"`
	FileType       string `dynamodbav:"FileType"`
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

// GetManifestsForDataset returns all manifests for a given dataset.
func GetManifestsForDataset(client *dynamodb.Client, manifestTableName string, datasetNodeId string) ([]ManifestTable, error) {

	queryInput := dynamodb.QueryInput{
		TableName:              aws.String(manifestTableName),
		IndexName:              aws.String("DatasetManifestIndex"),
		KeyConditionExpression: aws.String("DatasetNodeId = :datasetValue"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":datasetValue": &types.AttributeValueMemberS{Value: datasetNodeId},
		},
		Select: "ALL_ATTRIBUTES",
	}

	result, err := client.Query(context.Background(), &queryInput)
	if err != nil {
		return nil, err
	}

	items := []ManifestTable{}
	for _, item := range result.Items {
		manifest := ManifestTable{}
		err = attributevalue.UnmarshalMap(item, &manifest)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalMap: %v\n", err)
		}
		items = append(items, manifest)
	}

	return items, nil
}

// UpdateFileTableStatus updates the status of the file in the file-table dynamodb
func UpdateFileTableStatus(client *dynamodb.Client, tableName string, manifestId string, uploadId string, status manifestFile.Status) error {

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

// GetFilesForPath returns files in path for a manifest with optional filter.
func GetFilesForPath(client *dynamodb.Client, tableName string, manifestId string, path string, filter string,
	limit int32, startKey map[string]types.AttributeValue) (*dynamodb.QueryOutput, error) {

	queryInput := dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("PathIndex"),
		ExclusiveStartKey:         startKey,
		ExpressionAttributeNames:  nil,
		ExpressionAttributeValues: nil,
		FilterExpression:          aws.String(filter),
		KeyConditionExpression:    aws.String(fmt.Sprintf("partitionKeyName=%s AND sortKeyName=%s", manifestId, path)),
		Limit:                     &limit,
		Select:                    "ALL_ATTRIBUTES",
	}

	result, err := client.Query(context.Background(), &queryInput)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetManifestFile returns a manifest file from the ManifestFile Table.
func GetManifestFile(client *dynamodb.Client, tableName string, manifestId string, uploadId string) (*ManifestFileTable, error) {
	item := ManifestFileTable{}

	data, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ManifestId": &types.AttributeValueMemberS{Value: manifestId},
			"UploadId":   &types.AttributeValueMemberS{Value: uploadId},
		},
	})

	if err != nil {
		return &item, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return &item, fmt.Errorf("GetItem: ManifestFile not found.\n")
	}

	err = attributevalue.UnmarshalMap(data.Item, &item)
	if err != nil {
		return &item, fmt.Errorf("UnmarshalMap: %v\n", err)
	}

	return &item, nil
}
