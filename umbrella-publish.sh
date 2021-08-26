#!/bin/bash
set -xe
# remove the trailing release version from the mgmt-api branch.
TRAVIS_BRANCH=${TRAVIS_BRANCH%-*}

# Environment props
echo "UMBRELLA_REPO_PATH=${UMBRELLA_REPO_PATH}"
echo "VALUES_FILE=${VALUES_FILE}"
echo "TRAVIS_BRANCH=${TRAVIS_BRANCH}"
echo "DOCKER_IMAGE_NAME=${DOCKER_IMAGE_NAME}"
echo "BUILD_ID=${BUILD_ID}"
REPO_NAME=$(echo ${UMBRELLA_REPO_PATH} | sed 's/^.*\///' | sed 's/.git//')
WORKDIR=$(pwd)

# Loop incase there is another git upload that happened and merge would be needed to resolve
for ITER in {1..30}
  do
    # Sync the Umbrella Repo pulled in from Environment props
    echo "Sync Umbrella Repo"
    cd ${WORKDIR}
    UMBRELLA_ACCESS_REPO_URL=${UMBRELLA_REPO_PATH:0:8}${gitApiKey}@${UMBRELLA_REPO_PATH:8}
    git clone ${UMBRELLA_ACCESS_REPO_URL}
    cd ${REPO_NAME}
    if git branch -a | grep "${TRAVIS_BRANCH}"; then
      git checkout ${TRAVIS_BRANCH}
    else
      git checkout -b ${TRAVIS_BRANCH}
      git config --global user.email "hribld@us.ibm.com"
      git config --global user.name "hribld"
      git status
    fi

    # Insert container registry and image tag values
    echo "Update Docker Image Build"

    CURRENT_TAG=$(grep -m 1 "\stag: " ${VALUES_FILE} | awk '{print"tag: " $2}');

    sed -i -e "s~${CURRENT_TAG}~tag: ${BUILD_ID}~" ${VALUES_FILE} && rm -f ${VALUES_FILE}-e

    # Commit and push to umbrella repo
    echo "Commit and Push"
    cd ${WORKDIR}
    cd ${REPO_NAME}
    git config --global user.email "hribld@us.ibm.com"
    git config --global user.name "hribld"
    git status
    git diff ${VALUES_FILE}
    git add ${VALUES_FILE}

    if git status | grep "nothing to commit";then
      echo "nothing to commit, there was a problem, something should have changed.  Dying."
      cd ${WORKDIR}
      rm -rf ${REPO_NAME}
      exit 1
    else
      git commit -a -m "Auto Promotion to ${VALUES_FILE}, DOCKER_IMAGE_NAME=${DOCKER_IMAGE_NAME}, BUILD_NUMBER=${BUILD_ID}"
      if git push origin ${TRAVIS_BRANCH};then
        echo "push good, breaking"
        cd ${WORKDIR}
        rm -rf ${REPO_NAME}
        break
      else
        echo "push failed, start over"
        cd ${WORKDIR}
        rm -rvf ${REPO_NAME}
      fi
    fi
    if [[ ${ITER} -gt 29 ]]; then
      echo "failed to push the image to umbrella repo, exiting... $ITER"
      exit 1
    fi
done
