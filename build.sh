#!/usr/bin/env bash

BUILD_SOCIAL_NETWORK_COMMAND="docker build -f Dockerfile.socialnetwork -t andrey540/socialnetwork:v4 ."
BUILD_USER_COMMAND="docker build -f Dockerfile.user -t andrey540/user:v4 ."
BUILD_CONVERSATION_COMMAND="docker build -f Dockerfile.conversation -t andrey540/conversation:v4 ."

make

${BUILD_SOCIAL_NETWORK_COMMAND}
${BUILD_USER_COMMAND}
${BUILD_CONVERSATION_COMMAND}
