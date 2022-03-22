#!/usr/bin/env bash

BUILD_SOCIAL_NETWORK_COMMAND="docker build -f Dockerfile.socialnetwork -t andrey540/userapp:v1 ."

make

${BUILD_SOCIAL_NETWORK_COMMAND}
