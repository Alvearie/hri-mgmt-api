#!/bin/bash

make 2>&1 > unittest
rtn=$?
cat unittest
< unittest go-junit-report > unittest.xml
rm unittest
exit $rtn
