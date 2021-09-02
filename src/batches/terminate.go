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

type Terminate struct{}

func (Terminate) GetAction() string {
	return "terminate"
}

func (Terminate) CheckAuth(claims auth.HriClaims) error {
	// Only Integrators can call terminate
	if !claims.HasScope(auth.HriIntegrator) {
		return errors.New(fmt.Sprintf(auth.MsgIntegratorRoleRequired, "update"))
	}
	return nil
}

func (Terminate) GetUpdateScript(params map[string]interface{}, validator param.Validator, claims auth.HriClaims, logger *log.Logger) (map[string]interface{}, map[string]interface{}) {
	errResp := validator.ValidateOptional(
		params,
		param.Info{param.Metadata, reflect.Map},
	)
	if errResp != nil {
		logger.Printf("Bad input optional params: %s", errResp)
		return nil, errResp
	}
	metadata := params[param.Metadata]

	currentTime := time.Now().UTC()

	if metadata == nil {
		updateScript := fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.endDate = '%s';} else {ctx.op = 'none'}",
			status.Started, claims.Subject, status.Terminated, currentTime.Format(elastic.DateTimeFormat))

		updateRequest := map[string]interface{}{
			"script": map[string]interface{}{
				"source": updateScript,
			},
		}
		return updateRequest, nil
	} else {
		updateScript := fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.endDate = '%s'; ctx._source.metadata = params.metadata;} else {ctx.op = 'none'}",
			status.Started, claims.Subject, status.Terminated, currentTime.Format(elastic.DateTimeFormat))

		updateRequest := map[string]interface{}{
			"script": map[string]interface{}{
				"source": updateScript,
				"lang":   "painless",
				"params": map[string]interface{}{"metadata": metadata},
			},
		}
		return updateRequest, nil
	}
}
