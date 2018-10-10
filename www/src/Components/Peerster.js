import React, {Component} from 'react';
import {Col, Row} from 'react-bootstrap';
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

    render() {
        return (
            <Col>
                <Col md={2}>
                </Col>
                <Col md={6}>
                    <Row>
                        <MessagesWindow messages={this.state.messages}/>
                    </Row>
                    <Row>
                        <Chatbox onSend={this.sendMessage}/>
                    </Row>
                </Col>
                <Col md={4}>
                    <Row>
                        <IDBox identifier={this.state.identifier}/>
                    </Row>
                    <Row>
                        <PeersList peers={this.state.peers}/>
                        <PeerAdder onAdd={this.addPeer}/>
                    </Row>
                </Col>
            </Col>
        )
    }

    getGossiperIdentifier = async () => {
        this.getGossiper(
            '/identifier/',
            (body) => this.setState({identifier: body}),
        );
    };

    getGossiperPeers = async () => {
        this.getGossiper('/peers/', (body) => this.setState({peers: body}));
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