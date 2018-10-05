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
            messages: {},
            wants: {},
        };
        this.getGossiperIdentifier =
            this.getGossiperIdentifier.bind(this);
        this.getGossiperIdentifier();
        this.getGossiperPeers();

        this.buildWantsString = this.buildWantsString.bind(this);
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

    // [{"Peer":"249498","Messages":[{"Origin":"249498","ID":1,"Text":"Leo"},{"Origin":"249498","ID":2,"Text":"M0v3f4st"}]}]

    getGossiperMessages = async () => {
        this.getGossiper('/message/' + this.buildWantsString(), (body) => {
            if(body === null) {
                return
            }
            const messages = Object.assign({}, this.state.messages);
            const wants = Object.assign({}, this.state.wants);
            body.forEach((peer) => {
                if (!(peer['Peer'] in messages)) {
                    messages[peer['Peer']] = [];
                    wants[peer['Peer']] = 1;
                }
                if(peer['Messages'] === null) {
                    return
                }
                    peer['Messages'].forEach((msg) => {
                        if (msg['ID'] >= wants[peer['Peer']]) {
                            messages[peer['Peer']].push(msg['Text']);
                            wants[peer['Peer']] = msg['ID'] + 1;
                        }
                    })
            });
            this.setState({messages: messages, wants: wants});
        });
    };

    buildWantsString = () => {
        let ret = '';
        Object.keys(this.state.wants).forEach((peer) => {
            ret += peer + ':' + this.state.wants[peer] + ';';
        });
        if(ret.length > 0) {
            ret = ret.slice(0, -1);
            ret += '/';
        }
        console.log('wants string', ret);
        return ret;
    };

    getGossiper = async (api, callback) => {
        const request = endPoint + api;
        const response = await fetch(request);
        const body = await response.json();
        if (response.status !== 200) throw Error(body.message);
        callback(body);
    }
}