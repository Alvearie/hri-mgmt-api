package mongoApi

import (
	"math/rand"
	"time"

	"github.com/Alvearie/hri-mgmt-api/common/model"
)

const DateTimeFormat string = "2006-01-02T15:04:05Z"

func ConvertToJSON(tenantId string, docCount string, docsDeleted int) model.CreateTenantRequest {
	result := model.CreateTenantRequest{
		TenantId:     tenantId,
		Docs_count:   docCount,
		Docs_deleted: docsDeleted,
	}
	return result
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomString(length int) string {
	return StringWithCharset(length, charset)
}
