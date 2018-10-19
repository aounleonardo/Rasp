#!/usr/bin/env bash

go build
cd client
go build
cd handler
go build
cd ../..

trap cleanup SIGINT SIGTERM

./Peerster -UIPort=8080 -gossipAddr=0.0.0.0:5000 -name=Leo > A.out &

./client/handler/handler > H.out &
(cd www/; yarn start > ../R.out &)

echo "Initialization done"

function cleanup {
    pkill -f Peerster
    pkill -f handler
    kill $(lsof -t -i:3000)
    exit 0
}

while true
do
    sleep 10
done

