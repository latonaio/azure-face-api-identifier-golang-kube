package main

import (
	"azure-face-api-identifier-kube-golang/src/identifier"
	"azure-face-api-identifier-kube-golang/src/model"
	"context"
	"database/sql"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/Azure/go-autorest/autorest"
	"github.com/latonaio/golang-logging-library/logger"
	rabbitmq "github.com/latonaio/rabbitmq-golang-client"
)

var apiKey = os.Getenv("API_ACCESS_KEY")
var endpoint = os.Getenv("API_ENDPOINT")
var rabbitmqURL = os.Getenv("RABBITMQ_URL")
var queueOrigin = os.Getenv("QUEUE_ORIGIN")
var queueTo = os.Getenv("QUEUE_TO")
var serviceName = os.Getenv("SERVICE_NAME")
var MYSQL_DSN = os.Getenv("MYSQL_DSN")
var PersonGroupID = os.Getenv("PERSON_GROUP_ID")
var redisDSN = os.Getenv("REDIS_DSN")

func main() {
	logging := logger.NewLogger()
	var setPersonParam model.SetPersonParam
	setPersonParam.Ctx = context.Background()
	setPersonParam.Client = face.NewClient(endpoint)
	setPersonParam.Client.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.Pgc = face.NewPersonGroupClient(endpoint)
	setPersonParam.Pgc.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.Pgpc = face.NewPersonGroupPersonClient(endpoint)
	setPersonParam.Pgpc.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.PersonGroupID = PersonGroupID

	// SQL Connection
	db, err := sql.Open("mysql", MYSQL_DSN)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	logging.Info("connect mysql")

	// rabbitmq Connection
	rabbitmqClient, err := rabbitmq.NewRabbitmqClient(
		rabbitmqURL,
		[]string{queueOrigin},
		[]string{queueTo},
	)
	if err != nil {
		logging.Error("can't connect rabbitmq")
		return
	}

	defer rabbitmqClient.Close()
	iter, err := rabbitmqClient.Iterator()
	if err != nil {
		logging.Error("not working iterator")
		return
	}
	logging.Info("connect rabbitmq")

	logging.Info("start azure-face-api identifier")

	for message := range iter {
		identifier.AzureFaceApiIdentifier(message, &setPersonParam, db, rabbitmqClient, serviceName, queueTo, redisDSN)
		message.Success()

	}
}
