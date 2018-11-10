import React, {Component} from 'react';
import {Col, Grid, Row} from 'react-bootstrap';
import IDBox from "./IDBox";
import PeersList from "./PeersList";
import MessagesWindow from "./MessagesWindow";
import Chatbox from "./Chatbox";
import PeerAdder from "./PeerAdder";
import Chats from "./Chats";
import Toolbar from "./Toolbar";

const endPoint = "http://127.0.0.1:8000";

const maybeIndex =
    (index) => (index === undefined || index === null) ? 0 : index;
const maybeList = (list) => (list === undefined || list === null) ? [] : list;
const merge = (obj, key, value) => {
    let ret = {...obj};
    ret[key] = value;
    return ret;
};

export default class Peerster extends Component {
    constructor(props) {
        super(props);

        this.state = {
            identifier: "",
            peers: [],
            chats: [],
            messages: [],
            wants: 0,
            currentChat: "",
            unordered: {},
            ordered: {},
            unorderedIndex: {},
            orderedIndex: {},
        };
        this.getGossiperIdentifier();
        this.getGossiperPeers();
        setInterval(this.getGossiperPeers, 2000);

        this.getGossiperChats();
        setInterval(this.getGossiperChats, 2000);

        this.getGossiperMessages();
        setInterval(this.getGossiperMessages, 1000);

        this.getGossiperPrivates();
        setInterval(this.getGossiperPrivates, 1000);
    }

    style = {
        peerster: {
            backgroundColor: "dodgerblue",
        },
        messagesWindow: {
            overflowY: "auto",
            overflowX: "hidden",
            height: "calc(80vh - 200px)",
        },
        chatbox: {
            height: "calc(20vh + 20px)",
        },
        gossipInfo: {
            height: "calc(80vh - 200px)",
            width: "80%",
            paddingLeft: "8%",
        },
        peerAdder: {
            height: "calc(20vh + 20px)",
            width: "80%",
        },
    };

    render() {
        return (
            <Grid style={this.style.peerster}>
                <Row>
                    <Col md={8}>
                        <Row style={this.style.messagesWindow}>
                            <MessagesWindow
                                identifier={this.state.identifier}
                                messages={this.state.messages}
                                unordered={maybeList(
                                    this.state.unordered[this.state.currentChat]
                                )}
                                ordered={maybeList(
                                    this.state.ordered[this.state.currentChat]
                                )}
                                currentChat={this.state.currentChat}
                            />
                        </Row>
                        <Row style={this.style.chatbox}>
                            <Chatbox onSend={this.sendMessage}/>
                        </Row>
                    </Col>
                    <Col md={4}>
                        <Row style={this.style.gossipInfo}>
                            <Row>
                                <IDBox identifier={this.state.identifier}/>
                            </Row>
                            <Row>
                                <Chats
                                    current={this.state.currentChat}
                                    identifier={this.state.identifier}
                                    peers={this.state.chats}
                                    chatSelected={this.chatSelected}
                                />
                            </Row>
                            <Row>
                                <PeersList peers={this.state.peers}/>
                            </Row>
                        </Row>
                        <Row style={this.style.peerAdder}>
                            <PeerAdder onAdd={this.addPeer}/>
                        </Row>
                    </Col>
                </Row>
                <Row>
                    <Toolbar
                        shareFile={this.shareFile}
                        download={this.downloadFile}
                    />
                </Row>
            </Grid>
        )
    }

    getGossiperIdentifier = async () => {
        this.getGossiper(
            '/identifier/',
            (body) => this.setState({identifier: body}),
        );
    };

    getGossiperPeers = async () => {
        this.getGossiper('/peers/', (body) => this.setState(
            {peers: (body !== null) ? body : []}
        ));
    };

    getGossiperChats = async () => {
        this.getGossiper('/chats/', (body) => this.setState(
            {chats: (body !== null) ? body : []}
        ));
    };

    getGossiperMessages = async () => {
        this.getGossiper('/message/' + this.state.wants + '/', (body) => {
            if (body === null) {
                return
            }
            const startIndex = body["StartIndex"];
            const receivedMessages = body["Messages"];
            if (receivedMessages === null) {
                return
            }
            const toDrop = Math.max(this.state.wants - startIndex, 0);
            receivedMessages.slice(toDrop, receivedMessages.length);
            const newMessages = [
                ...this.state.messages,
                ...receivedMessages,
            ];
            const nextID = newMessages.length;

            this.setState({messages: newMessages, wants: nextID})
        });
    };

    getGossiperPrivates = async () => {
        if (this.state.currentChat === "") {
            return
        }
        this.getGossiper(
            '/pm/' +
            this.state.currentChat +
            '/' +
            maybeIndex(this.state.unorderedIndex[this.state.currentChat]) +
            '/' +
            maybeIndex(this.state.orderedIndex[this.state.currentChat]) +
            '/',
            (body) => {
                if (body === null) {
                    return
                }
                const unorderedIndex = maybeIndex(body["UnorderedIndex"]);
                const orderedIndex = maybeIndex(body["OrderedIndex"]);
                const receivedUnordered = maybeList(body["Unordered"]);
                const receivedOrdered = maybeList(body["Ordered"]);
                const partner = body["Partner"];

                let newUnordered = [...maybeList(
                    this.state.unordered[partner],
                )];
                let newUnorderedIndex = maybeIndex(
                    this.state.unorderedIndex[partner],
                );
                let newOrdered = [...maybeList(
                    this.state.ordered[partner],
                )];
                let newOrderedIndex = maybeIndex(
                    this.state.orderedIndex[partner],
                );

                if (receivedUnordered !== null) {
                    const toDrop =
                        Math.max(newUnorderedIndex - unorderedIndex, 0);
                    receivedUnordered.slice(toDrop, receivedUnordered.length);
                    newUnordered = [
                        ...newUnordered,
                        ...receivedUnordered,
                    ];
                    newUnorderedIndex = newUnordered.length;
                }

                if (receivedOrdered !== null) {
                    const toDrop = Math.max(newOrderedIndex - orderedIndex, 0);
                    receivedOrdered.slice(toDrop, receivedOrdered.length);
                    newOrdered = [
                        ...newOrdered,
                        ...receivedOrdered,
                    ];
                    newOrderedIndex = newOrdered.length;
                }

                this.setState({
                    unordered: merge(
                        this.state.unordered,
                        partner,
                        newUnordered,
                    ),
                    ordered: merge(
                        this.state.ordered,
                        partner,
                        newOrdered,
                    ),
                    unorderedIndex: merge(
                        this.state.unorderedIndex,
                        partner,
                        newUnorderedIndex,
                    ),
                    orderedIndex: merge(
                        this.state.orderedIndex,
                        partner,
                        newOrderedIndex,
                    ),
                });
            },
        );
    };

    getGossiper = async (api, callback) => {
        const request = endPoint + api;
        const response = await fetch(request);
        const body = await response.json();
        if (response.status !== 200) throw Error(body.message);
        callback(body);
    };

    sendMessage = async (message) => {
        const details = this.getMessageDetails(message);
        const request = endPoint + details.api;
        await fetch(request, {
            method: 'post',
            body: details.body,
        })
            .then(res => res.json())
            .then(res => (res === false) ?
                console.log("Error occurred while posting") :
                console.log(`Message ${message} sent.`));
    };

    shareFile = async (file, callback) => {
        const data = new FormData();
        const name = file.name;
        data.append("file", file, name);
        const request = endPoint + '/share-file/';
        await fetch(request, {
            method: 'post',
            body: data,
        })
            .then(res => res.json())
            .then(res => {
                if (res.hasOwnProperty("Metakey")) {
                    this.sendMessage(
                        `I just shared the file: "${file.name}"\n` +
                        `with Metahash: "${res["Metakey"]}"`
                    );
                    callback(res["Metakey"]);
                }
            });
    };

    downloadFile = async (metakey, filename, callback) => {
        const request = endPoint + '/download-file/';
        if (this.state.currentChat === "") {
            callback(false, "choose a peer first");
            return
        }
        await fetch(request, {
            method: 'post',
            body: JSON.stringify({
                Metakey: metakey,
                Filename: filename,
                Origin: this.state.currentChat,
            }),
        })
            .then(res => res.json())
            .then(res => callback(res["Success"], ""))
    };

    getMessageDetails = (message) => {
        return (this.state.currentChat === "") ? {
            api: '/message/',
            body: JSON.stringify(message),
        } : {
            api: '/pm/',
            body: JSON.stringify({
                Contents: message,
                Destination: this.state.currentChat,
            })
        };
    };

    addPeer = async (address, port) => {
        const request = endPoint + '/peers/';
        fetch(request, {
            method: 'post',
            body: JSON.stringify({address: address, port: port}),
        })
            .then(res => res.json())
            .then(res => (res === false) ?
                console.log("Error occurred while adding peer") :
                console.log(`Peer ${address}:${port} added.`));
    };

    chatSelected = (peer) => {
        if (peer !== this.state.peer) {
            if (peer === "") {
                this.getGossiperMessages();
            } else {
                this.getGossiperPrivates();
            }
        }
        this.setState({currentChat: peer});
    }
}