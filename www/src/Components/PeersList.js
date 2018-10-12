import React, {Component} from 'react';
import {Col, Label, Row} from 'react-bootstrap';

export default class PeersList extends Component {
    render() {
        return <Col>
            {this.props.peers.map((peer) => this.createRow(peer))}
        </Col>

    }

    createRow = (peer) => {
        return <Row key={peer}>
            <Label>{peer}</Label>
        </Row>
    }
}