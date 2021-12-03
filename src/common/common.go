package common

import (
	"azure-face-api-identifier-kube-golang/src/model"
	"azure-face-api-identifier-kube-golang/src/redis"
	"fmt"
	"strconv"

	"github.com/latonaio/golang-logging-library/logger"
	rabbitmq "github.com/latonaio/rabbitmq-golang-client"
)

var logging = logger.NewLogger()

// common package は、main.goで扱う具体的な処理の集まりです。
func SetStatus(identifyFaceStatus model.IdentifyFaceStatus, guestID int, serviceName string, imgPath string, redisKey string) model.IdentifyResult {
	var data model.IdentifyResult
	switch {
	case identifyFaceStatus.Customer == "existing" && *identifyFaceStatus.Confifendence > 0.60 && guestID != 0:
		data.ConnectionKey = "response"
		data.Result = true
		data.RedisKey = redisKey
		data.FilePath = imgPath
		data.Person = identifyFaceStatus
		data.GuestID = guestID
		data.Status = "existing"

	case identifyFaceStatus.Customer == "new":
		data.ConnectionKey = "response"
		data.Result = true
		data.RedisKey = redisKey
		data.FilePath = imgPath
		data.Status = "new"

	case identifyFaceStatus.Customer == "error":
		data.ConnectionKey = "response"
		data.Result = false
		data.RedisKey = redisKey
		data.Microservice = serviceName
	}

	return data
}

func GetDataFromOriginQueue(message rabbitmq.RabbitmqMessage, setPersonParam *model.SetPersonParam) {
	for index, value := range message.Data() {
		fmt.Println("value:", value)
		if index == "image_path" {
			setPersonParam.ImgPath = value.(string)
		}
		if index == "guest_key" {
			// guestKeyはrabbitmq-client-golang の仕様で float64が渡される
			guestKey := strconv.FormatFloat(value.(float64), 'f', 0, 64)
			// guestKeyの例: 1638411461478
			setPersonParam.RedisKey = guestKey
		}
	}
}

func InsertDataToRedis(data model.IdentifyResult, redisDSN string) {
	redisKey := data.RedisKey
	customer := data.Status
	result := data.Result
	filePath := data.FilePath

	insertRedis := new(model.InsertToRedis)

	c := redis.Connection(redisDSN)
	defer c.Close()

	switch {
	case (result) && customer == "existing":
		insertRedis.Status = "success"
		insertRedis.Customer = customer
		insertRedis.GuestID = data.GuestID
	case (result) && customer == "new":
		insertRedis.Status = "success"
		insertRedis.Customer = customer
		insertRedis.ImagePath = filePath
	case (!result):
		insertRedis.Status = "failed"
		insertRedis.FailedMS = "MICROSERVICE"
	}

	// データの登録(Redis: SET key value)
	res_set := redis.SetData(redisKey, insertRedis, c)

	logging.Info("set data to redis:", res_set)
}
