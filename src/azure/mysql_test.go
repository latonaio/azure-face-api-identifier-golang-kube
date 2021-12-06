package azure

import (
	"azure-face-api-identifier-golang-kube/src/model"
	"database/sql"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/gofrs/uuid"
)

func TestGetGuestID(t *testing.T) {

	result := prepare()

	// SQLの準備
	db, err := sql.Open("mysql", "root:root@(192.168.XXX.XXX:30000)/Omotebako")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close() // 関数がリターンする直前に呼び出される

	testGuestID := 12
	personID := result.PersonID.String()

	// 初期化処理
	var args []interface{} = []interface{}{testGuestID, result.PersonID.String()}
	db.Query("INSERT INTO guest (guest_id, face_id_azure) VALUES(?,?)", args...)

	guestID := GetGuestID(db, personID)
	notHaveGuestID := GetGuestID(db, "invalid args")

	// 後処理
	db.Query("DELETE FROM guest WHERE guest_id = ?", testGuestID)
	if err != nil {
		panic(err.Error())
	}

	if guestID == testGuestID {
		t.Logf("success")
	} else {
		t.Errorf("something not wrong")
	}

	if notHaveGuestID == 0 {
		t.Logf("success")
	} else {
		t.Errorf("something not wrong")
	}
}

func prepare() *model.IdentifyFaceStatus {
	// 画像の取得
	image, err := os.Open("../../face/initial.jpeg")
	if err != nil {
		panic(err)
	}
	defer image.Close()

	setPersonParam.Img = image

	// FaceIDを取得
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

	return result

}
