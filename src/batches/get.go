/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	paramGteDate string = "gteDate"
	paramLteDate string = "lteDate"
	paramSize    string = "size"
	paramFrom    string = "from"
)

func checkForInjection(str string) error {
	if strings.ContainsAny(str, `"[]{}`) {
		return errors.New("query parameters may not contain these characters: \"[]{}")
	}
	return nil
}

func appendTerm(params map[string]interface{}, param string, clauses *[]map[string]interface{}) error {
	value, ok := params[param].(string)
	if ok {
		if err := checkForInjection(value); err != nil {
			return err
		}
		clause := map[string]interface{}{
			"term": map[string]interface{}{
				param: value,
			},
		}
		*clauses = append(*clauses, clause)
	}
	return nil
}

func appendRange(params map[string]interface{}, paramName string, gteParam string, lteParam string, clauses *[]map[string]interface{}) error {
	gteValue, gteOk := params[gteParam].(string)
	lteValue, lteOk := params[lteParam].(string)

	if !gteOk && !lteOk {
		return nil
	}

	rangeClauses := map[string]interface{}{}
	if gteOk {
		if err := checkForInjection(gteValue); err != nil {
			return err
		}
		rangeClauses["gte"] = gteValue
	}
	if lteOk {
		if err := checkForInjection(lteValue); err != nil {
			return err
		}
		rangeClauses["lte"] = lteValue
	}

	clause := map[string]interface{}{
		"range": map[string]interface{}{
			paramName: rangeClauses,
		},
	}
	*clauses = append(*clauses, clause)
	return nil
}

func Get(params map[string]interface{}, claims auth.HriClaims, client *elasticsearch.Client) map[string]interface{} {
	logger := log.New(os.Stdout, "batches/get: ", log.Llongfile)

	if !claims.HasScope(auth.HriConsumer) && !claims.HasScope(auth.HriIntegrator) {
		errMsg := auth.MsgAccessTokenMissingScopes
		logger.Println(errMsg)
		return response.Error(http.StatusUnauthorized, errMsg)
	}

	tenantId, err := path.ExtractParam(params, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	mustClauses := make([]map[string]interface{}, 0, 4) //at most 4 query restrictions
	if err := appendTerm(params, param.Name, &mustClauses); err != nil {
		return response.Error(http.StatusBadRequest, err.Error())
	}
	if err := appendTerm(params, param.Status, &mustClauses); err != nil {
		return response.Error(http.StatusBadRequest, err.Error())
	}
	if err := appendRange(params, param.StartDate, paramGteDate, paramLteDate, &mustClauses); err != nil {
		return response.Error(http.StatusBadRequest, err.Error())
	}
	// If only the HriIntegrator role is present, filter results to batches it owns
	if !claims.HasScope(auth.HriConsumer) && claims.HasScope(auth.HriIntegrator) {
		clause := map[string]interface{}{
			"term": map[string]interface{}{
				param.IntegratorId: claims.Subject,
			},
		}
		mustClauses = append(mustClauses, clause)
	}

	var buf *bytes.Buffer
	if len(mustClauses) > 0 {
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": mustClauses,
				},
			},
		}
		buf, err = elastic.EncodeQueryBody(query)
		if err != nil {
			msg := fmt.Sprintf("Error encoding Elastic query: %s", err.Error())
			logger.Println(msg)
			return response.Error(http.StatusInternalServerError, msg)
		}
		logger.Printf("query: %v\n", query)
	} else {
		buf = &bytes.Buffer{}
	}

	var size = 10
	if params[paramSize] != nil {
		size, err = strconv.Atoi(params[paramSize].(string))
		if err != nil {
			msg := fmt.Sprintf("Error parsing 'size' parameter: %s", err.Error())
			logger.Println(msg)
			return response.Error(http.StatusBadRequest, msg)
		}
	}

	var from = 0
	if params[paramFrom] != nil {
		from, err = strconv.Atoi(params[paramFrom].(string))
		if err != nil {
			msg := fmt.Sprintf("Error parsing 'from' parameter: %s", err.Error())
			logger.Println(msg)
			return response.Error(http.StatusBadRequest, msg)
		}
	}

	index := elastic.IndexFromTenantId(tenantId)

	// Perform the search request.
	res, err := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithBody(buf),
		client.Search.WithSize(size),
		client.Search.WithFrom(from),
		client.Search.WithTrackTotalHits(true),
	)

	body, errResp := elastic.DecodeBody(res, err, tenantId, logger)
	if errResp != nil {
		// check for parse exceptions
		if body != nil {
			if body["error"].(map[string]interface{})["type"] == "search_phase_execution_exception" {
				causes, ok := body["error"].(map[string]interface{})["root_cause"].([]interface{})
				if !ok {
					logger.Println("unable to decode error.root_cause")
					return errResp
				}
				for _, value := range causes {
					entry := value.(map[string]interface{})
					if entry["type"].(string) == "parse_exception" {
						return response.Error(http.StatusBadRequest, fmt.Sprintf("%v: %v", entry["type"], entry["reason"]))
					}
				}
			}
		}

		return errResp
	}

	hits := body["hits"].(map[string]interface{})["hits"].([]interface{})
	for i, entry := range hits {
		hits[i] = EsDocToBatch(entry.(map[string]interface{}))
	}

	return response.Success(http.StatusOK, map[string]interface{}{
		"total":   body["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64),
		"results": hits,
	})
}
