package azure

import (
	"azure-face-api-identifier-kube-golang/src/model"
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/latonaio/golang-logging-library/logger"
)

var logging = logger.NewLogger()

// 顔を検知
func DetectFacesFromImage(setPersonParam *model.SetPersonParam) (face.ListDetectedFace, error) {

	var err error
	// DetectWithStreamから取得するパラメータを設定
	attributes := []face.AttributeType{}
	returnFaceID := true
	returnRecognitionModel := false
	returnFaceLandmarks := false
	detectSingleFaces, dErr := setPersonParam.Client.DetectWithStream(
		setPersonParam.Ctx,
		setPersonParam.Img,
		&returnFaceID,
		&returnFaceLandmarks,
		attributes,
		face.Recognition03,
		&returnRecognitionModel,
		face.Detection02)

	if dErr != nil {
		logging.Error("failed getting rectangle")
		err = dErr
	}

	return detectSingleFaces, err
}

// 登録した画像がAzureAPI上に存在し、登録した顔と検知した顔とで同一かを確認
func IdentityFromRegisterdFace(setPersonParam *model.SetPersonParam, identifyRequest face.IdentifyRequest) (*model.IdentifyFaceStatus, []face.IdentifyCandidate, error) {

	resultPersonList := new(model.IdentifyFaceStatus)

	// 最も似ているPersonIDを取得
	identifyResult, err := setPersonParam.Client.Identify(setPersonParam.Ctx, identifyRequest)
	if err != nil {
		logging.Error("can't identify", err)
		resultPersonList.Customer = "error"
	}

	value := *identifyResult.Value
	candidateList := *value[0].Candidates

	if len(candidateList) > 0 {
		mostSimilarPersonID := candidateList[0].PersonID

		// PersonGroupIDから画像の登録リストを取得
		var personLists, _ = setPersonParam.Pgpc.List(setPersonParam.Ctx, setPersonParam.PersonGroupID, "", nil)

		// 登録した顔と検知した顔とを比較
		for _, v := range *personLists.Value {
			if v.PersonID.String() == mostSimilarPersonID.String() {
				logging.Info("this customer is existing")
				resultPersonList.Customer = "existing"
				resultPersonList.PersonID = candidateList[0].PersonID
				resultPersonList.Confifendence = candidateList[0].Confidence
			}
		}
	}
	if len(candidateList) == 0 {
		resultPersonList.Customer = "new"
		logging.Info("this customer is new")
	}
	return resultPersonList, candidateList, err
}

// Azure に顔画像をトレーニングしてもらう
func TrainPersonImage(ctx context.Context, pgc face.PersonGroupClient, personGroupID string) (string, error) {
	pgc.Train(ctx, personGroupID)
	var trainState string
	var err error

	for {
		isTrainStatus, _ := pgc.GetTrainingStatus(ctx, personGroupID)

		if isTrainStatus.Status == face.TrainingStatusTypeSucceeded {
			trainState = "success"
			break
		}
		if isTrainStatus.Status == face.TrainingStatusTypeFailed {
			trainState = "failed"
			break
		}
	}

	if trainState == "failed" {
		logging.Error("failed to train")
		err = fmt.Errorf("failed to train")
		return trainState, err
	}

	return trainState, err
}
