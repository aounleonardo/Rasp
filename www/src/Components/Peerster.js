import React, {Component} from 'react';
import {Col, Row} from 'react-bootstrap';
import IDBox from "./IDBox";

export default class Peerster extends Component {
    constructor(props) {
        super(props);
        this.state = {
            identifier: '',
        };
        this.requestGossiperIdentifier =
            this.requestGossiperIdentifier.bind(this);
        this.requestGossiperIdentifier();
    }

    render() {
        return(
            <Col>
                <Col md={8}>
                    Chat Box
                </Col>
                <Col md={4}>
                    <Row>
                        Peers
                    </Row>
                    <Row>
                        <IDBox identifier={this.state.identifier}/>
                    </Row>
                </Col>
            </Col>
        )
    }

    requestGossiperIdentifier = async () => {
        const request = 'http://127.0.0.1:8000/identifier';
        const response = await fetch(request);
        const body = await response.text();
        if (response.status !== 200) throw Error(body.message);
        this.setState({identifier: body});
    }
}