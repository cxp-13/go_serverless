package user

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"strconv"
	"strings"
)

var (
	ErrorFailedToFetchRecord      = "failed to fetch record"
	ErrorFailedToUnmarshalRecord  = "failed to unmarshal record"
	ErrorFailedToFetchRecords     = "failed to fetch records"
	ErrorFailedToUnmarshalRecords = "failed to unmarshal records"
	ErrorCouldNotMarshalItem      = "could not marshal item"
	ErrorCouldNotPutItem          = "could not put item"
	ErrorFailedToDeleteItem       = "failed to delete item"
	ErrorFailedToUpdateItem       = "failed to update item"
)

type User struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func FetchUser(email, tableName string, dynaClient *dynamodb.DynamoDB) (*User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	}

	result, err := dynaClient.GetItem(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}
	if result.Item == nil {
		return nil, nil // User not found
	}

	item := new(User)
	if err := dynamodbattribute.UnmarshalMap(result.Item, item); err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}
	return item, nil
}

func FetchUsers(tableName string, dynaClient *dynamodb.DynamoDB) ([]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := dynaClient.Scan(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecords)
	}

	items := []User{}
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecords)
	}
	return items, nil
}

func CreateUser(user User, tableName string, dynaClient *dynamodb.DynamoDB) (*User, error) {
	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotPutItem)
	}
	return &user, nil
}

func UpdateUser(email string, updates map[string]interface{}, tableName string, dynaClient *dynamodb.DynamoDB) error {
	updateExpressions := []string{}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{}
	//expressionAttributeNames := map[string]*string{}

	for key, value := range updates {
		//updateExpressions = append(updateExpressions, "#"+key+" = :"+key)
		updateExpressions = append(updateExpressions, key+" = :"+key)
		//expressionAttributeNames["#"+key] = aws.String(key)

		// Dynamically set the correct type for the AttributeValue
		switch v := value.(type) {
		case string:
			expressionAttributeValues[":"+key] = &dynamodb.AttributeValue{S: aws.String(v)}
		case int:
			expressionAttributeValues[":"+key] = &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(v))}
		case bool:
			expressionAttributeValues[":"+key] = &dynamodb.AttributeValue{BOOL: aws.Bool(v)}
		// Add other types as needed
		default:
			return errors.New("unsupported attribute value type")
		}
	}

	input := &dynamodb.UpdateItemInput{
		TableName:        aws.String(tableName),
		Key:              map[string]*dynamodb.AttributeValue{"email": {S: aws.String(email)}},
		UpdateExpression: aws.String("set " + strings.Join(updateExpressions, ", ")),
		//ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	}

	_, err := dynaClient.UpdateItem(input)
	if err != nil {
		return errors.New(ErrorFailedToUpdateItem)
	}
	return nil
}

func DeleteUser(email, tableName string, dynaClient *dynamodb.DynamoDB) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	}

	_, err := dynaClient.DeleteItem(input)
	if err != nil {
		return errors.New(ErrorFailedToDeleteItem)
	}
	return nil
}

func UnhandledMethod() {
	// This is a placeholder for handling unsupported methods.
}
