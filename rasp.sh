#!/usr/bin/env bash

PORT=8000

go build
cd client
go build
cd server
go build
cd ../..

trap cleanup SIGINT SIGTERM

./Peerster -UIPort=8080 -gossipAddr=127.0.0.1:5000 -name=A -peers=127.0.0.1:5001 -rtimer=10 > A.out &
./Peerster -UIPort=8081 -gossipAddr=127.0.0.1:5001 -name=B -peers=127.0.0.1:5000 -rtimer=10 > B.out &
./Peerster -UIPort=8082 -gossipAddr=127.0.0.1:5002 -name=C -peers=127.0.0.1:5000 -rtimer=10 > C.out &
./Peerster -UIPort=8083 -gossipAddr=127.0.0.1:5003 -name=D -peers=127.0.0.1:5002 -rtimer=10 > D.out &

bash -c "exec -a peersterServerA ./client/server/server -port=$PORT -gossiper=8080  > HA.out &"
bash -c "exec -a peersterServerB ./client/server/server -port=8001 -gossiper=8081  > HB.out &"
bash -c "exec -a peersterServerC ./client/server/server -port=8002 -gossiper=8082  > HB.out &"
bash -c "exec -a peersterServerD ./client/server/server -port=8003 -gossiper=8083  > HB.out &"

(cd www/; bash -c "exec -a peersterGui yarn start > ../R.out &")
sleep 5
open http://localhost:3000/8001
open http://localhost:3000/8002
open http://localhost:3000/8003

echo "Initialization done"

function cleanup {
    pkill -f Peerster
    pkill -f peersterServerA
    pkill -f peersterServerB
    pkill -f peersterServerC
    pkill -f peersterServerD
    pkill -f peersterGui
    exit 0
}

while true
do
    sleep 10
done
