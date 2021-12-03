package identifier

import (
	"azure-face-api-identifier-kube-golang/src/azure"
	"azure-face-api-identifier-kube-golang/src/common"
	"azure-face-api-identifier-kube-golang/src/model"
	"database/sql"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/gofrs/uuid"
	"github.com/latonaio/golang-logging-library/logger"
	rabbitmq "github.com/latonaio/rabbitmq-golang-client"
)

func AzureFaceApiIdentifier(
	message rabbitmq.RabbitmqMessage,
	setPersonParam *model.SetPersonParam,
	db *sql.DB, rabbitmqClient *rabbitmq.RabbitmqClient,
	serviceName string,
	queueTo string,
	redisDSN string) {

	logging := logger.NewLogger()

	// メッセージキューからデータを取得
	common.GetDataFromOriginQueue(message, setPersonParam)

	// 取得したデータから渡された画像パスで画像を開く
	img, err := os.Open(setPersonParam.ImgPath)
	if err != nil {
		panic(err)
	}
	defer img.Close()
	logging.Info("open image")

	setPersonParam.Img = img

	// 画像をもとに顔を検知
	faces, err := azure.DetectFacesFromImage(setPersonParam)
	if err != nil {
		panic(err)
	}

	if len(*faces.Value) == 0 {
		panic("face is not detected")
	}

	// FaceIDを取得
	dFaces := *faces.Value
	faceID := *dFaces[0].FaceID
	logging.Info("faceID is :", faceID)

	// FaceID と PersonGroupID を Identify に用いる構造体にセット
	var identifyBody face.IdentifyRequest
	identifyBody.FaceIds = &[]uuid.UUID{faceID}
	identifyBody.PersonGroupID = &setPersonParam.PersonGroupID

	// 登録した画像がAzureAPI上に存在するか、また、登録した顔と検知した顔とで同一かを確認
	identifyFaceStatus, candidateList, err := azure.IdentityFromRegisterdFace(setPersonParam, identifyBody)
	if err != nil {
		panic(err)
	}

	var guestID int = 0
	if identifyFaceStatus.Customer == "existing" {
		guestID = azure.GetGuestID(db, identifyFaceStatus.PersonID.String())
	}

	// 既存客・新規客・例外かを判定
	data := common.SetStatus(*identifyFaceStatus, guestID, serviceName, setPersonParam.ImgPath, setPersonParam.RedisKey)

	var payload = map[string]interface{}{
		"imagePath": setPersonParam.ImgPath,
		"faceId":    faceID,
		"responseData": map[string][]face.IdentifyCandidate{
			"candidates": candidateList,
		},
	}

	// 送信先のキューに送信
	if err := rabbitmqClient.Send(queueTo, payload); err != nil {
		log.Printf("error: %v", err)
	}
	logging.Info("send to %v", queueTo)

	// redisに書き込む
	common.InsertDataToRedis(data, redisDSN)
	logging.Info("inserted to redis")
}
