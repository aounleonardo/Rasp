import React, {Component} from 'react';
import {Col, Row} from 'react-bootstrap';
import IDBox from "./IDBox";
import PeersList from "./PeersList";
import MessagesWindow from "./MessagesWindow";

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
    }

    render() {
        return (
            <Col>
                <Col md={2}>
                </Col>
                <Col md={6}>
                    <MessagesWindow messages={this.state.messages}/>
                </Col>
                <Col md={4}>
                    <Row>
                        <IDBox identifier={this.state.identifier}/>
                    </Row>
                    <Row>
                        <PeersList peers={this.state.peers}/>
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
                ... this.state.messages,
                ... receivedMessages,
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
    }
}