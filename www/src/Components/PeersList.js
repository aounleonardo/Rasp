import React, {Component} from 'react';
import {Col, Label} from 'react-bootstrap';

export default class PeersList extends Component {
    render() {
        return <Col>
            {this.props.peers.sort().map(
                (peer) => this.createRow(peer),
            )}
        </Col>

    }

    createRow = (peer) => {
        return <h4 key={peer}>
            <Label bsStyle={"info"}>{peer}</Label>
        </h4>
    }
}