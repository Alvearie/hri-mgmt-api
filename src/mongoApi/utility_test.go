package mongoApi

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	randomString := RandomString(4)
	assert.NotEmpty(t, randomString)

}

func TestConvertToJSON(t *testing.T) {
	expected := model.CreateTenantRequest{
		TenantId:     "test_tenanats",
		Docs_count:   "0",
		Docs_deleted: 2,
	}
	actual := ConvertToJSON("test_tenanats", "0", 2)
	assert.NotEmpty(t, actual)
	assert.Equal(t, expected, actual)

}
