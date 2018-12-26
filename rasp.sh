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

./client/server/server -port=$PORT -gossiper=8080  > HA.out &
./client/server/server -port=8001 -gossiper=8081  > HB.out &

(cd www/; yarn start > ../R.out &)
sleep 3
open http://localhost:3000/8001

echo "Initialization done"

function cleanup {
    pkill -f Peerster
    kill $(lsof -t -i:$PORT)
    kill $(lsof -t -i:8001)
    pkill $(lsof -t -i:3000)
    exit 0
}

while true
do
    sleep 10
done

