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
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"log"
	"reflect"
	"time"
)

type SendComplete struct{}

func (SendComplete) GetAction() string {
	return "sendComplete"
}

func (SendComplete) CheckAuth(claims auth.HriClaims) error {
	// Only Integrators can call sendComplete
	if !claims.HasScope(auth.HriIntegrator) {
		return errors.New(fmt.Sprintf(auth.MsgIntegratorRoleRequired, "update"))
	}
	return nil
}

func (SendComplete) GetUpdateScript(params map[string]interface{}, validator param.Validator, claims auth.HriClaims, logger *log.Logger) (map[string]interface{}, map[string]interface{}) {
	errResp := validator.Validate(
		params,
		// golang receives numeric JSON values as Float64
		param.Info{param.Validation, reflect.Bool},
	)
	if errResp != nil {
		logger.Printf("Bad input params: %s", errResp)
		return nil, errResp
	}

	errResp = validator.ValidateOptional(
		params,
		param.Info{param.ExpectedRecordCount, reflect.Float64},
		param.Info{param.RecordCount, reflect.Float64},
		param.Info{param.Metadata, reflect.Map},
	)
	if errResp != nil {
		logger.Printf("Bad input optional params: %s", errResp)
		return nil, errResp
	}

	var expectedRecordCount int
	if _, present := params[param.ExpectedRecordCount]; present {
		expectedRecordCount = int(params[param.ExpectedRecordCount].(float64))
	} else if _, present := params[param.RecordCount]; present {
		expectedRecordCount = int(params[param.RecordCount].(float64))
	} else {
		return nil, response.MissingParams(param.ExpectedRecordCount)
	}
	validation := params[param.Validation].(bool)

	metadata := params[param.Metadata]

	var updateScript string
	if validation {
		// When validation is enabled
		// - change the status to 'sendCompleted'
		// - set the record count

		if metadata == nil {
			updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.expectedRecordCount = %d;} else {ctx.op = 'none'}",
				status.Started, claims.Subject, status.SendCompleted, expectedRecordCount)
		} else {
			updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.expectedRecordCount = %d; ctx._source.metadata = params.metadata;} else {ctx.op = 'none'}",
				status.Started, claims.Subject, status.SendCompleted, expectedRecordCount)
		}

	} else {
		// When validation is not enabled
		// - change the status to 'completed'
		// - set the record count
		// - set the end date
		currentTime := time.Now().UTC()

		if metadata == nil {
			updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.expectedRecordCount = %d; ctx._source.endDate = '%s';} else {ctx.op = 'none'}",
				status.Started, claims.Subject, status.Completed, expectedRecordCount, currentTime.Format(elastic.DateTimeFormat))
		} else {
			updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.expectedRecordCount = %d; ctx._source.endDate = '%s'; ctx._source.metadata = params.metadata;} else {ctx.op = 'none'}",
				status.Started, claims.Subject, status.Completed, expectedRecordCount, currentTime.Format(elastic.DateTimeFormat))
		}
	}

	var updateRequest = map[string]interface{}{}

	if metadata == nil {
		updateRequest = map[string]interface{}{
			"script": map[string]interface{}{
				"source": updateScript,
			},
		}
	} else {
		updateRequest = map[string]interface{}{
			"script": map[string]interface{}{
				"source": updateScript,
				"lang":   "painless",
				"params": map[string]interface{}{"metadata": metadata},
			},
		}
	}

	return updateRequest, nil
}
