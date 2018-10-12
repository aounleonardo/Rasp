import React, {Component} from 'react';
import {Col, Grid, Row} from 'react-bootstrap';
import IDBox from "./IDBox";
import PeersList from "./PeersList";
import MessagesWindow from "./MessagesWindow";
import Chatbox from "./Chatbox";
import PeerAdder from "./PeerAdder";

const endPoint = 'http://127.0.0.1:8000';

export default class Peerster extends Component {
    constructor(props) {
        super(props);
        this.state = {
            identifier: '',
            peers: [],
            messages: [],
            wants: 0,
        };
        this.getGossiperIdentifier =
            this.getGossiperIdentifier.bind(this);
        this.getGossiperIdentifier();
        this.getGossiperPeers();
        setInterval(this.getGossiperPeers, 5000);

        this.getGossiperMessages = this.getGossiperMessages.bind(this);
        this.getGossiperMessages();
        setInterval(this.getGossiperMessages, 3000);

        this.sendMessage = this.sendMessage.bind(this);
        this.addPeer = this.addPeer.bind(this);
    }

    style = {
        messagesWindow: {
            backgroundColor: "dodgerblue",
            overflowY: "auto",
            overflowX: "hidden",
            height: "calc(80vh - 200px)",
        },
        chatbox: {
            height: "calc(20vh + 20px)",
            backgroundColor: "dodgerblue",
        },
        gossipInfo: {
            height: "calc(80vh - 200px)",
            width: "80%",
            backgroundColor: "dodgerblue",
            paddingLeft: "8%",
        },
        peerAdder: {
            height: "calc(20vh + 20px)",
            width: "80%",
            backgroundColor: "dodgerblue",
        },
    };

    render() {
        return (
            <Grid>
                <Col md={8}>
                    <Row style={this.style.messagesWindow}>
                        <MessagesWindow identifier={this.state.identifier} messages={this.state.messages}/>
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
                            <PeersList peers={this.state.peers}/>
                        </Row>
                    </Row>
                    <Row style={this.style.peerAdder}>
                        <PeerAdder onAdd={this.addPeer}/>
                    </Row>
                </Col>
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
        this.getGossiper('/peers/', (body) => this.setState({peers: (body !== null)? body : []}));
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

    getGossiper = async (api, callback) => {
        const request = endPoint + api;
        const response = await fetch(request);
        const body = await response.json();
        if (response.status !== 200) throw Error(body.message);
        callback(body);
    };

    sendMessage = async (message) => {
        const request = endPoint + '/message/';
        fetch(request, {
            method: 'post',
            body: JSON.stringify(message),
        })
            .then(res => res.json())
            .then(res => (res === false) ?
                // TODO this is not working, probably because it is a bool
                console.log("Error occurred while posting") :
                console.log(`Message ${message} sent.`));
    };

    addPeer = async (address, port) => {
        const request = endPoint + '/peers/';
        fetch(request, {
            method: 'post',
            body: JSON.stringify({address: address, port: port}),
        })
            .then(res => res.json())
            .then(res => (res === false) ?
                // TODO this is not working, probably because it is a bool
                console.log("Error occurred while adding peer") :
                console.log(`Peer ${address}:${port} added.`));
    }
}