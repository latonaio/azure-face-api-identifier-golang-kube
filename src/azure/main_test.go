package azure

import (
	"azure-face-api-identifier-golang-kube/src/model"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/Azure/go-autorest/autorest"
)

var apiKey = "xxxxx"
var endpoint = "xxxx"
var setPersonParam model.SetPersonParam

/* テストのキャッシュを削除するコマンド
*  go clean -testcache
 */
func TestMain(m *testing.M) {

	setPersonParam.Ctx = context.Background()
	setPersonParam.Client = face.NewClient(endpoint)
	setPersonParam.Client.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.Pgc = face.NewPersonGroupClient(endpoint)
	setPersonParam.Pgc.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.Pgpc = face.NewPersonGroupPersonClient(endpoint)
	setPersonParam.Pgpc.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.PersonGroupID = "test-group"

	constructor(setPersonParam)
	code := m.Run()
	deconstructor(setPersonParam)
	os.Exit(code)
}

func constructor(setPersonParam model.SetPersonParam) {
	var metadata face.MetaDataContract
	metadata.Name = &setPersonParam.PersonGroupID
	metadata.RecognitionModel = face.Recognition03
	var body face.NameAndUserDataContract
	guestID := "hoge"
	body.Name = &guestID

	image, err := os.Open("../../face/initial.jpeg")
	if err != nil {
		panic(err)
	}
	defer image.Close()

	// persongroupの作成
	setPersonParam.Pgc.Create(setPersonParam.Ctx, setPersonParam.PersonGroupID, metadata)

	// personGroup personの作成
	person, err := setPersonParam.Pgpc.Create(setPersonParam.Ctx, setPersonParam.PersonGroupID, body)
	if err != nil {
		panic(err)
	}
	personID := person.PersonID

	// 画像をpersonIDに割り当てる
	targetRC := []int32{124, 82, 73, 73} // left, top, width, height
	setPersonParam.Pgpc.AddFaceFromStream(setPersonParam.Ctx, setPersonParam.PersonGroupID, *personID, image, "", targetRC, face.Detection02)

	// 画像をトレーニング
	_, terr := TrainPersonImage(setPersonParam.Ctx, setPersonParam.Pgc, setPersonParam.PersonGroupID)
	if terr != nil {
		panic(err)
	}

	var lists, _ = setPersonParam.Pgpc.List(setPersonParam.Ctx, setPersonParam.PersonGroupID, "", nil)
	for _, list := range *lists.Value {
		fmt.Printf("🔥 %v に登録されている personID と Name \n", setPersonParam.PersonGroupID)
		fmt.Println("🔥 personID", list.PersonID)
		fmt.Println("🔥 Name", *list.Name)
	}
}

func deconstructor(setPersonParam model.SetPersonParam) {
	setPersonParam.Pgc.Delete(setPersonParam.Ctx, setPersonParam.PersonGroupID)
	fmt.Printf("💦 personGroupID: %v を削除しました。\n", setPersonParam.PersonGroupID)
}
