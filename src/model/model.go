package model

import (
	"context"
	"image"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/cognitiveservices/face"
	"github.com/gofrs/uuid"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

type SetPersonParam struct {
	Ctx           context.Context
	Client        face.Client                  // 顔の検出、類似の検索、および検証の例に使用されるクライアント
	Pgpc          face.PersonGroupPersonClient // PersonGroupにPersonを追加する際に使用されるクライアント
	Pgc           face.PersonGroupClient       // PersonGroupに使用されるクライアント
	Img           *os.File
	ImgPath       string
	PersonGroupID string
	GuestID       float64
	RedisKey      string
}

type IdentifyFaceStatus struct {
	Customer      string
	PersonID      *uuid.UUID
	Confifendence *float64
}

type IdentifyResult struct {
	ConnectionKey string
	Result        bool
	RedisKey      string
	FilePath      string
	Person        IdentifyFaceStatus
	GuestID       int
	Status        string // 新規(new) or 既存(existing)
	Microservice  string
}

type InsertToRedis struct {
	Status    string
	Customer  string
	GuestID   int
	FailedMS  string
	ImagePath string
}
