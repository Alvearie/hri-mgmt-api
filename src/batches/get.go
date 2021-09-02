/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"net/http"
)

const defaultSize = 10
const defaultFrom = 0

func Get(requestId string, params model.GetBatch, claims auth.HriClaims, client *elasticsearch.Client) (int, interface{}) {
	prefix := "batches/get"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Get")

	// Data Integrators and Consumers can use this endpoint, so either scope allows access
	if !claims.HasScope(auth.HriConsumer) && !claims.HasScope(auth.HriIntegrator) {
		errMsg := auth.MsgAccessTokenMissingScopes
		logger.Errorln(errMsg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, errMsg)
	}

	return get(requestId, params, false, &claims, client, logger)
}

func GetNoAuth(requestId string, params model.GetBatch, _ auth.HriClaims, client *elasticsearch.Client) (int, interface{}) {
	prefix := "batches/getNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Get (No Auth)")

	var noAuthFlag = true
	return get(requestId, params, noAuthFlag, nil, client, logger)
}

func get(requestId string, params model.GetBatch, noAuthFlag bool, claims *auth.HriClaims,
	client *elasticsearch.Client, logger logrus.FieldLogger) (int, interface{}) {

	var buf *bytes.Buffer
	var err error

	query := buildQuery(params, claims)
	if query != nil {
		buf, err = elastic.EncodeQueryBody(query)
		if err != nil {
			msg := fmt.Sprintf("Error encoding Elastic query: %s", err.Error())
			logger.Errorln(msg)
			return http.StatusInternalServerError, response.NewErrorDetail(requestId, msg)
		}
		logger.Infof("query: %v\n", query)
	} else {
		buf = &bytes.Buffer{}
	}

	tenantId := params.TenantId
	index := elastic.IndexFromTenantId(tenantId)
	size, from := getClientSearchParams(params)
	// Perform the search request.
	res, err := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithBody(buf),
		client.Search.WithSize(size),
		client.Search.WithFrom(from),
		client.Search.WithTrackTotalHits(true),
	)

	body, elasticErr := elastic.DecodeBody(res, err)
	if elasticErr != nil {
		if elasticErr.Code == http.StatusUnauthorized {
			return http.StatusInternalServerError, elasticErr.LogAndBuildErrorDetail(requestId, logger, "Get batch failed")
		}
		return elasticErr.Code, elasticErr.LogAndBuildErrorDetail(requestId, logger, "Get batch failed")
	}

	hits := body["hits"].(map[string]interface{})["hits"].([]interface{})
	for i, entry := range hits {
		hits[i] = EsDocToBatch(entry.(map[string]interface{}))
	}

	return http.StatusOK, map[string]interface{}{
		"total":   body["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64),
		"results": hits,
	}
}

func buildQuery(params model.GetBatch, claims *auth.HriClaims) map[string]interface{} {
	noAuthFlag := false // (default -> Auth Enabled)
	if claims == nil {
		noAuthFlag = true
	}

	clauses := make([]map[string]interface{}, 0, 4) // at most 4 query restrictions
	if params.Name != nil {
		appendClause(&clauses, "term", param.Name, *params.Name)
	}
	if params.Status != nil {
		appendClause(&clauses, "term", param.Status, *params.Status)
	}

	if params.GteDate != nil || params.LteDate != nil {
		rangeClauses := map[string]interface{}{}
		if params.GteDate != nil {
			rangeClauses["gte"] = *params.GteDate
		}
		if params.LteDate != nil {
			rangeClauses["lte"] = *params.LteDate
		}
		appendClause(&clauses, "range", param.StartDate, rangeClauses)
	}

	// If only the HriIntegrator role is present, filter results to batches it owns
	if !noAuthFlag && !claims.HasScope(auth.HriConsumer) &&
		claims.HasScope(auth.HriIntegrator) {
		appendClause(&clauses, "term", param.IntegratorId, claims.Subject)
	}

	if len(clauses) == 0 {
		return nil
	}
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": clauses,
			},
		},
	}
	return query
}

func appendClause(clauses *[]map[string]interface{}, clauseType string, paramName string, paramVal interface{}) {
	clause := map[string]interface{}{
		clauseType: map[string]interface{}{
			paramName: paramVal,
		},
	}
	*clauses = append(*clauses, clause)
}

func getClientSearchParams(params model.GetBatch) (int, int) {
	var size int
	var from int
	if params.Size == nil {
		size = defaultSize
	} else {
		size = *params.Size
	}
	if params.From == nil {
		from = defaultFrom
	} else {
		from = *params.From
	}
	return size, from
}
