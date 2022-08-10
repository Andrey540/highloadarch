#!/usr/bin/env bash

BUILD_SOCIAL_NETWORK_COMMAND="docker build -f Dockerfile.socialnetwork -t andrey540/socialnetwork:v13 ."
BUILD_USER_COMMAND="docker build -f Dockerfile.user -t andrey540/user:v13 ."
BUILD_CONVERSATION_COMMAND="docker build -f Dockerfile.conversation -t andrey540/conversation:v13 ."
BUILD_POST_COMMAND="docker build -f Dockerfile.post -t andrey540/post:v13 ."
BUILD_COUNTER_COMMAND="docker build -f Dockerfile.counter -t andrey540/counter:v13 ."

make

${BUILD_SOCIAL_NETWORK_COMMAND}
${BUILD_USER_COMMAND}
${BUILD_CONVERSATION_COMMAND}
${BUILD_POST_COMMAND}
${BUILD_COUNTER_COMMAND}
