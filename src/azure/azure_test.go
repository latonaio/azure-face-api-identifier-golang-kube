package azure

import (
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/gofrs/uuid"
)

// 顔が取得できることを確認
func TestGetFace(t *testing.T) {
	image, err := os.Open("../../face/initial.jpeg")
	if err != nil {
		panic(err)
	}
	defer image.Close()

	setPersonParam.Img = image

	faces, err := DetectFacesFromImage(&setPersonParam)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(*faces.Value) == 0 {
		t.Errorf("image is not faces")
	}

	if len(*faces.Value) > 0 {
		t.Logf("success")
	}
}

// 顔が取得できないことを確認
func TestGetNotFace(t *testing.T) {

	image, err := os.Open("../../face/not_face.jpeg")
	if err != nil {
		panic(err)
	}
	defer image.Close()

	setPersonParam.Img = image

	faces, err := DetectFacesFromImage(&setPersonParam)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(*faces.Value) > 0 {
		t.Errorf("image has faces")
	}

	t.Logf("success")
}

// 登録した画像がAzureAPI上に存在し、登録した顔と検知した顔とで同一かを確認
func TestIdentityFromRegisterdFace(t *testing.T) {

	image, err := os.Open("../../face/initial.jpeg")
	if err != nil {
		panic(err)
	}
	defer image.Close()

	setPersonParam.Img = image

	// 顔を検知し、FaceIDを受け取る。その後FaceIDがどの登録されているIDと似ているかをチェック
	faces, err := DetectFacesFromImage(&setPersonParam)
	if err != nil {
		panic(err)
	}
	dFaces := *faces.Value
	faceID := *dFaces[0].FaceID
	personGroupID := "test-group"

	var identifyBody face.IdentifyRequest
	identifyBody.FaceIds = &[]uuid.UUID{faceID}
	identifyBody.PersonGroupID = &personGroupID
	result, _, err := IdentityFromRegisterdFace(&setPersonParam, identifyBody)
	if err != nil {
		panic(err)
	}

	if result.Customer == "existing" || result.Customer == "new" {
		t.Logf("success")
	} else {
		t.Errorf("occur missing detection")
	}

}
