/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

/*func TestCreate(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	requestId := "request_id_1"
	tenantId := "tenant-123_"

	elasticErrMsg := "elasticErrMsg"

	testCases := []struct {
		name      string
		transport *test.FakeTransport
		expected  interface{}
	}{
		{
			name: "bad-response",
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches", tenantId),
				test.ElasticCall{
					ResponseErr: errors.New(elasticErrMsg),
				},
			),
			expected: response.NewErrorDetail(requestId, fmt.Sprintf("Unable to create new tenant [%s]: [500] elasticsearch client error: %s", tenantId, elasticErrMsg)),
		},
		{
			name: "good-request",
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches", tenantId),
				test.ElasticCall{
					ResponseBody: fmt.Sprintf(`{"acknowledged":true,"shards_acknowledged":true,"index":"%s-batches"}`, tenantId),
				},
			),
			expected: map[string]interface{}{"tenantId": tenantId},
		},
	}

	for _, tc := range testCases {
		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			if _, actual := Create(requestId, tenantId, client); !reflect.DeepEqual(tc.expected, actual) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tc.expected, actual))
			}
			tc.transport.VerifyCalls()
		})
	}
}*/
