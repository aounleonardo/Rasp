import React, {Component} from 'react';
import {Col, Row, Label} from 'react-bootstrap';

export default class MessagesWindow extends Component {
    render() {
        return(
            <Col>
                {Object.keys(this.props.messages).map((key) => this.createRows(key, this.props.messages[key]))}
            </Col>
        )
    }

    createRows = (peer, msgs) => {
        return <Row key={peer}>{msgs.map((msg) => <Row key={msg}><Label>{peer}: {msg}</Label></Row>)}</Row>
    }
}