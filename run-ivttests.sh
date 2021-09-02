#!/usr/bin/env bash
# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

echo 'Run IVT Tests'
rspec test/spec --tag ~@broken --format documentation --format RspecJunitFormatter --out ivttest.xml
