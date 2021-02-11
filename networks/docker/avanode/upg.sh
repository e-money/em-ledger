#!/bin/sh

##
## Once images are re-built, 
## and keystore exported/imported
## upgrade network gradually
## so that the chain is preserved 
##

# curl -X POST --data '{
#     "jsonrpc":"2.0",
#     "id"     :1,
#     "method" :"keystore.exportUser",
#     "params" :{
#         "username":"",
#         "password":""
#     }
# }' -H 'content-type:application/json;' 127.0.0.1:9650/ext/keystore | jq

# curl -X POST --data '{
#     "jsonrpc":"2.0",
#     "id"     :1,
#     "method" :"keystore.importUser",
#     "params" :{
#         "username":"",
#         "password":"",
#         "user"    :"...xxxxxx"
#     }
# }' -H 'content-type:application/json;' 127.0.0.1:9650/ext/keystore | jq

# exit at first error
# echo each command and expand
set -ex

for i in 5; do
    docker-compose rm -v -s -f emavanode$i
    echo sleeping 3 seconds for node to come down...
    sleep 3
    sudo rm -rf /var/lib/docker/volumes/avanode_avalanchevol$i/_data/
    sudo mkdir -p /var/lib/docker/volumes/avanode_avalanchevol$i/_data
    docker-compose up -d emavanode$i
    echo sleeping 4 seconds for new node to catch up...
    sleep 4
done
