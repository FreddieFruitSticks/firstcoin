#!/bin/bash

FIRSTCOIN_ECR_REPO=$(aws ecr describe-repositories | jq -r '.repositories' | jq -r '.[1].repositoryName')
IMAGES=$(docker images -aq)
echo "tagging $i ..."
docker tag $(echo "$IMAGES $FIRSTCOIN_ECR_REPO")
