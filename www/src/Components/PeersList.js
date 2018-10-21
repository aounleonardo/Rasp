import React, {Component} from 'react';
import {Col, Label} from 'react-bootstrap';

export default class PeersList extends Component {
    render() {
        return <Col>
            <h4>
                <Label bsStyle={"primary"}>Peers</Label>
            </h4>
            {this.props.peers.sort().map(
                (peer) => this.createRow(peer),
            )}
        </Col>

    }

    createRow = (peer) => {
        return <h5 key={peer}>
            <Label bsStyle={"info"}>{peer}</Label>
        </h5>
    }
}