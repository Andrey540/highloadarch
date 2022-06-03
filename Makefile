

# List of command package names,
#  each one builds from Go package at 'cmd/$NAME' to executable at 'bin/$NAME'
APP_CMD_NAMES = \
	socialnetwork \
	conversation \
	post \
	user

# List of proto files with API definition,
#  used to generate client/server Go code and REST API proxy (both Go code and Swagger docs)
APP_PROTO_FILES = \
	pkg/common/api/conversation.proto \
	pkg/common/api/post.proto \
	pkg/common/api/user.proto \

# Contains common make targets, including 'build', 'test' and 'check'
include make/rules.mk
