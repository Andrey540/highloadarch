#!/usr/bin/env bash

BUILD_SOCIAL_NETWORK_COMMAND="docker build -f Dockerfile.socialnetwork -t andrey540/socialnetwork:v11 ."
BUILD_USER_COMMAND="docker build -f Dockerfile.user -t andrey540/user:v11 ."
BUILD_CONVERSATION_COMMAND="docker build -f Dockerfile.conversation -t andrey540/conversation:v11 ."
BUILD_POST_COMMAND="docker build -f Dockerfile.post -t andrey540/post:v11 ."
BUILD_COUNTER_COMMAND="docker build -f Dockerfile.counter -t andrey540/counter:v11 ."

make

${BUILD_SOCIAL_NETWORK_COMMAND}
${BUILD_USER_COMMAND}
${BUILD_CONVERSATION_COMMAND}
${BUILD_POST_COMMAND}
${BUILD_COUNTER_COMMAND}
