#!/usr/bin/env bash

# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

set -e

echo 'Run IVT Deploy Tests'
rspec test/spec/hri_management_api_deploy_spec.rb --tag ~@broken --format documentation --format RspecJunitFormatter --out test/ivt_test_results/ivttest_deploy.xml

echo 'Run IVT Tests Without Validation'
rspec test/spec/hri_management_api_no_validation_spec.rb --tag ~@broken --format documentation --format RspecJunitFormatter --out test/ivt_test_results/ivttest_no_validation.xml

echo 'Run IVT Tests With Validation'
rspec test/spec/hri_management_api_validation_spec.rb --tag ~@broken --format documentation --format RspecJunitFormatter --out test/ivt_test_results/ivttest_validation.xml