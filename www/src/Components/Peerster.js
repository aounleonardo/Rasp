import React, {Component} from 'react';
import {Col, Row} from 'react-bootstrap';
import IDBox from "./IDBox";
import PeersList from "./PeersList";

const endPoint = 'http://127.0.0.1:8000';

export default class Peerster extends Component {
    constructor(props) {
        super(props);
        this.state = {
            identifier: '',
            peers: [],
        };
        this.requestGossiperIdentifier =
            this.requestGossiperIdentifier.bind(this);
        this.requestGossiperIdentifier();
        this.requestGossiperPeers()
    }

    render() {
        return(
            <Col>
                <Col md={8}>
                    Chat Box
                </Col>
                <Col md={4}>
                    <Row>
                        <PeersList peers={this.state.peers}/>
                    </Row>
                    <Row>
                        <IDBox identifier={this.state.identifier}/>
                    </Row>
                </Col>
            </Col>
        )
    }

    requestGossiperIdentifier = async () => {
        const request = endPoint + '/identifier/';
        const response = await fetch(request);
        const body = await response.json();
        if (response.status !== 200) throw Error(body.message);
        this.setState({identifier: body});
    };

    requestGossiperPeers = async () => {
      const request = endPoint + '/peers/';
      const response = await fetch(request);
      const body = await response.json();
      if (response.status !== 200) throw Error(body.message);
      this.setState({peers: body})
    }
}