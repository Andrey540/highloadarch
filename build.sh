#!/usr/bin/env bash

BUILD_SOCIAL_NETWORK_COMMAND="docker build -f Dockerfile.socialnetwork -t andrey540/socialnetwork:v1 ."

make

${BUILD_SOCIAL_NETWORK_COMMAND}
