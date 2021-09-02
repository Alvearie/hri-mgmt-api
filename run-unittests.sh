#!/bin/bash

# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

make test 2>&1 > unittest
rtn=$?
cat unittest
< unittest go-junit-report > unittest.xml
rm unittest
exit $rtn
