scan_name=$([ ! -z "$TRAVIS_BRANCH" ] && [ ! -z "$TRAVIS_BUILD_NUMBER" ] &&  echo "TRAVIS-CI-$TRAVIS_BRANCH-$TRAVIS_BUILD_NUMBER" ||  echo "$USER-$(date +'%Y-%b-%d-%H-%M-%S')")

wget -O ../SAClientUtil.zip $ASOC_CLI_URL
unzip ../SAClientUtil.zip -d ../
find .. -type d | sort | grep SAClientUtil | while read folder; do mv $folder ../SAClientUtil; break; done;

cp -r $GOPATH/pkg/mod dependencies
../SAClientUtil/bin/appscan.sh api_login -P $ASOC_KEY_SECRET -u $ASOC_KEY_ID -persist

echo $SCAN_NAME
../SAClientUtil/bin/appscan.sh prepare -d $(pwd) -l $(pwd) -n appscan -v
../SAClientUtil/bin/appscan.sh queue_analysis -a $ASOC_APP_ID -f ./appscan.irx -n $scan_name
