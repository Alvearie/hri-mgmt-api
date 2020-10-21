#!/usr/bin/env bash

echo 'Run IVT Tests'
rspec test/spec --tag ~@broken --format documentation --format RspecJunitFormatter --out ivttest.xml
