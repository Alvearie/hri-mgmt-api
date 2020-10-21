/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
$util.insights.set('service_name', 'wh-hri-mgmt-api');
// Each IBM functions namespace has a unique API endpoint URL, which can be
// retrieved from the IBM Cloud dashboard. When creating a monitor in New Relic
// you will need to include the endpoint in the string below.
const HEALTH_CHECK_URL = '<api_endpoint>/healthcheck';

const assert = require('assert');

function evaluateHealth(err, response, body) {
    if (err) {
        assert.fail('Unable to contact endpoint.', err.message);
    }

    assert.ok(response.statusCode == 200, 'Expected a 200 OK response, got ' + response.statusCode);
}

const options = {
    url: HEALTH_CHECK_URL,
    headers: {
        // API Key required for accessing HRI-API namespace; secret must be created
        // in the New Relic web application
        'X-IBM-Client-Id': $secure.HRI_MGMT_API_KEY,
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    }
}
$http.get(options, evaluateHealth);
