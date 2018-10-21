import React, {Component} from 'react';
import {Col, Grid, Row} from 'react-bootstrap';
import IDBox from "./IDBox";
import PeersList from "./PeersList";
import MessagesWindow from "./MessagesWindow";
import Chatbox from "./Chatbox";
import PeerAdder from "./PeerAdder";
import Chats from "./Chats";
import Toolbar from "./Toolbar";

const endPoint = 'http://127.0.0.1:8000';

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
            unordered: [],
            ordered: [],
            unorderedIndex: 0,
            orderedIndex: 0,
        };
        this.getGossiperIdentifier =
            this.getGossiperIdentifier.bind(this);
        this.getGossiperIdentifier();
        this.getGossiperPeers();
        setInterval(this.getGossiperPeers, 5000);

        this.getGossiperChats();
        setInterval(this.getGossiperChats, 5000);

        this.getGossiperMessages = this.getGossiperMessages.bind(this);
        this.getGossiperMessages();
        setInterval(this.getGossiperMessages, 3000);

        this.getGossiperPrivates = this.getGossiperPrivates.bind(this);
        this.getGossiperPrivates();
        setInterval(this.getGossiperPrivates, 3000);

        this.sendMessage = this.sendMessage.bind(this);
        this.addPeer = this.addPeer.bind(this);
        this.chatSelected = this.chatSelected.bind(this);
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
                                unordered={this.state.unordered}
                                ordered={this.state.ordered}
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
                    <Toolbar/>
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
            this.state.unorderedIndex +
            '/' +
            this.state.orderedIndex +
            '/',
            (body) => {
                if (body === null) {
                    return
                }
                const unorderedIndex = body["UnorderedIndex"];
                const orderedIndex = body["OrderedIndex"];
                const receivedUnordered = body["Unordered"];
                const receivedOrdered = body["Ordered"];
                let newUnordered = [...this.state.unordered];
                let newUnorderedIndex = this.state.unorderedIndex;
                let newOrdered = [...this.state.ordered];
                let newOrderedIndex = this.state.orderedIndex;

                if (receivedUnordered !== null) {
                    const toDrop =
                        Math.max(this.state.unorderedIndex - unorderedIndex, 0);
                    receivedUnordered.slice(toDrop, receivedUnordered.length);
                    newUnordered = [
                        ...this.state.unordered,
                        ...receivedUnordered,
                    ];
                    newUnorderedIndex = newUnordered.length;
                }

                if (receivedOrdered !== null) {
                    const toDrop =
                        Math.max(this.state.orderedIndex - orderedIndex, 0);
                    receivedOrdered.slice(toDrop, receivedOrdered.length);
                    newOrdered = [
                        ...this.state.ordered,
                        ...receivedOrdered,
                    ];
                    newOrderedIndex = newOrdered.length;
                }

                this.setState({
                    unordered: newUnordered,
                    ordered: newOrdered,
                    unorderedIndex: newUnorderedIndex,
                    orderedIndex: newOrderedIndex,
                })
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
        fetch(request, {
            method: 'post',
            body: details.body,
        })
            .then(res => res.json())
            .then(res => (res === false) ?
                console.log("Error occurred while posting") :
                console.log(`Message ${message} sent.`));
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
        if(peer !== this.state.peer) {
            if(peer === "") {
                this.getGossiperMessages();
            } else {
                this.getGossiperPrivates();
            }
        }
        this.setState({currentChat: peer});
    }
}