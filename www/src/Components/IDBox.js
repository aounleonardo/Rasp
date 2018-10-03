import React, {Component} from 'react';
import {Col, Row, Label} from 'react-bootstrap';

export default class IDBox extends Component {
    render() {
        return(
            <Col>
                <Row>
                    <Label>Gossiper:{this.props.identifier}</Label>
                </Row>
            </Col>
        )
    }
}