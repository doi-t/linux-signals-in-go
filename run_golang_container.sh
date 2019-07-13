#!/bin/bash
set -ex
IMAGE='golang'
TARGET=/go

docker pull ${IMAGE}:latest

cat <<-EOF
Commands you might need:
$ docker rm $$(docker ps -aq) # Cleanup exitted containers
$ docker exec -it dev_${IMAGE} /bin/bash # Login to a running container
EOF

docker run \
  -it \
  --name dev_${IMAGE} \
  --mount type=bind,source="$(pwd)",target=${TARGET} \
  ${IMAGE}:latest \
  /bin/bash
