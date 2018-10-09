import React, {Component} from 'react';
import {Col, Row, Label} from 'react-bootstrap';

export default class MessagesWindow extends Component {
    render() {
        return(
            <Col>
                {this.props.messages.map((message) => this.createRow(message))}
            </Col>
        )
    }

    createRow = (msg) => {
        return <Row key={`${msg["Origin"]}:${msg["ID"]}`}>
            <Label>{msg["Origin"]}: {msg["Text"]}</Label>
        </Row>
    }
}