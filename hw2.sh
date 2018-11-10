#!/usr/bin/env bash

go build
cd client
go build
cd handler
go build
cd ../..

mkdir -p test/_Downloads/_Chunks/
mkdir test/_SharedFiles/
cp Peerster test/
mkdir test/client/
cp client/client test/client/

trap cleanup SIGINT SIGTERM

./Peerster -UIPort=8080 -gossipAddr=127.0.0.1:5000 -name=A -peers=127.0.0.1:5002 -rtimer=10 > A.out &
(cd test/; ./Peerster -UIPort=8081 -gossipAddr=127.0.0.1:5001 -name=B -peers=127.0.0.1:5002,127.0.0.1:5003 -rtimer=10 > B.out &)
./Peerster -UIPort=8082 -gossipAddr=127.0.0.1:5002 -name=C -peers=127.0.0.1:5001 -rtimer=10 > C.out &
./Peerster -UIPort=8083 -gossipAddr=127.0.0.1:5003 -name=D -peers=127.0.0.1:5001 -rtimer=10 > D.out &

./client/handler/handler > H.out &

(cd test/; ../client/handler/handler -port=8001 -gossiper=8081 > HB.out &)
(cd www/; yarn start > ../R.out &)

echo "Initialization done"

echo "I'm Batman" > _SharedFiles/A.txt
out=$(./client/client -UIPort=8080 -file="A.txt")
./client/client -UIPort=8080 -dest="B" -msg="A.txt $out"

sleep 2
cd test/

out=$(./client/client -UIPort=8081 -dest="A" -file="A.txt" -request=$out)
out=$(cat _Downloads/A.txt)
./client/client -UIPort=8081 -dest="A" -msg="read: A.txt $out"

echo "And I'm Robin" > _SharedFiles/B.txt
out=$(./client/client -UIPort=8081 -file="B.txt")
./client/client -UIPort=8081 -dest="A" -msg="B.txt $out"
cd ../


function cleanup {
    pkill -f Peerster
    pkill -f handler
    kill $(lsof -t -i:3000)

    rm -r _SharedFiles/*
    rm -r _Downloads/*
    mkdir _Downloads/_Chunks/

    rm -r test/

    exit 0
}

while true
do
    sleep 10
done

