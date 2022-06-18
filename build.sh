#!/usr/bin/env bash

BUILD_SOCIAL_NETWORK_COMMAND="docker build -f Dockerfile.socialnetwork -t andrey540/socialnetwork:v10 ."
BUILD_USER_COMMAND="docker build -f Dockerfile.user -t andrey540/user:v10 ."
BUILD_CONVERSATION_COMMAND="docker build -f Dockerfile.conversation -t andrey540/conversation:v10 ."
BUILD_POST_COMMAND="docker build -f Dockerfile.post -t andrey540/post:v10 ."
BUILD_COUNTER_COMMAND="docker build -f Dockerfile.counter -t andrey540/counter:v10 ."

make

${BUILD_SOCIAL_NETWORK_COMMAND}
${BUILD_USER_COMMAND}
${BUILD_CONVERSATION_COMMAND}
${BUILD_POST_COMMAND}
${BUILD_COUNTER_COMMAND}
