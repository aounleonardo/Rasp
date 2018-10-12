import React, {Component} from 'react';
import {Label} from 'react-bootstrap';

export default class IDBox extends Component {
    render() {
        return <h1>
            <Label bsStyle={"primary"}>{this.props.identifier}</Label>
        </h1>

    }
}