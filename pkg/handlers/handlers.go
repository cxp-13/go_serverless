package handlers

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"go-serverless/pkg/user"
	"go-serverless/pkg/validators"
	"net/http"
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
	ErrorMethodNotAllowed         = "method not allowed"
)

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

func GetUser(req events.APIGatewayProxyRequest, tableName string, dynaClient *dynamodb.DynamoDB) (*events.APIGatewayProxyResponse, error) {
	email := req.QueryStringParameters["email"]
	if len(email) > 0 {
		result, err := user.FetchUser(email, tableName, dynaClient)
		if err != nil {
			return apiResponse(http.StatusBadGateway, ErrorBody{aws.String(err.Error())})
		}
		return apiResponse(http.StatusOK, result)
	}
	result, err := user.FetchUsers(tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusBadGateway, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusOK, result)
}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient *dynamodb.DynamoDB) (*events.APIGatewayProxyResponse, error) {
	var curUser user.User
	err := json.Unmarshal([]byte(req.Body), &curUser)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String(ErrorCouldNotMarshalItem)})
	}
	//validate email
	valid := validators.IsEmailValid(curUser.Email)

	if !valid {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String("invalid email")})
	}

	result, err := user.CreateUser(curUser, tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusBadGateway, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusCreated, result)
}

func UpdateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient *dynamodb.DynamoDB) (*events.APIGatewayProxyResponse, error) {
	email := req.QueryStringParameters["email"]
	if len(email) == 0 {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String("missing email")})
	}

	var updates map[string]interface{}
	err := json.Unmarshal([]byte(req.Body), &updates)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String("invalid request body")})
	}

	err = user.UpdateUser(email, updates, tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusBadGateway, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusOK, "successfully updated user")
}

func DeleteUser(req events.APIGatewayProxyRequest, tableName string, dynaClient *dynamodb.DynamoDB) (*events.APIGatewayProxyResponse, error) {
	email := req.QueryStringParameters["email"]
	if len(email) == 0 {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String("missing email")})
	}

	// 执行删除操作
	err := user.DeleteUser(email, tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusBadGateway, ErrorBody{aws.String("failed to delete user")})
	}

	return apiResponse(http.StatusOK, "successfully deleted user")
}

func UnhandledMethod() (*events.APIGatewayProxyResponse, error) {
	return apiResponse(http.StatusMethodNotAllowed, ErrorMethodNotAllowed)

}
