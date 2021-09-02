/*
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"log"
	"reflect"
	"time"
)

type ProcessingComplete struct{}

const ProcessingCompleteAction string = "processingComplete"

func (ProcessingComplete) GetAction() string {
	return ProcessingCompleteAction
}

func (ProcessingComplete) CheckAuth(claims auth.HriClaims) error {
	// Only internal code can call processing_complete
	if !claims.HasScope(auth.HriInternal) {
		return errors.New(fmt.Sprintf(auth.MsgInternalRoleRequired, "processing complete"))
	}
	return nil
}

func (ProcessingComplete) GetUpdateScript(params map[string]interface{}, validator param.Validator, _ auth.HriClaims, logger *log.Logger) (map[string]interface{}, map[string]interface{}) {
	errResp := validator.Validate(
		params,
		// golang receives numeric JSON values as Float64
		param.Info{param.ActualRecordCount, reflect.Float64},
		param.Info{param.InvalidRecordCount, reflect.Float64},
	)
	if errResp != nil {
		logger.Printf("Bad input params: %s", errResp)
		return nil, errResp
	}
	actualRecordCount := int(params[param.ActualRecordCount].(float64))
	invalidRecordCount := int(params[param.InvalidRecordCount].(float64))
	currentTime := time.Now().UTC()

	updateScript := fmt.Sprintf("if (ctx._source.status == '%s') {ctx._source.status = '%s'; ctx._source.actualRecordCount = %d; ctx._source.invalidRecordCount = %d; ctx._source.endDate = '%s';} else {ctx.op = 'none'}",
		status.SendCompleted, status.Completed, actualRecordCount, invalidRecordCount, currentTime.Format(elastic.DateTimeFormat))

	updateRequest := map[string]interface{}{
		"script": map[string]interface{}{
			"source": updateScript,
		},
	}
	return updateRequest, nil
}
