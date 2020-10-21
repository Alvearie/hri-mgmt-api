# index-templates

These JSON files represent ElasticSearch index templates which determine how records are stored and indexed.

**NOTE: Index templates define a default mapping for indexes based on their name. When creating new index templates, be mindful of the patterns used in other templates to avoid collisions and unpredictable indexing behavior.**

## Submitting Index Templates to Elastic

Index templates can be created directly through the Elastic REST API. For example, the following `cURL` command creates a `batches` template based on the content of `batches.json`:
```
CURL_CA_BUNDLE=/path/to/certificate curl -X PUT <elastic_endpoint>/_template/batches \
  -u admin:<password> \
  -H 'Content-Type: application/json' \
  -d '@batches.json'
```
