import React, {Component} from 'react';
import {FormControl} from 'react-bootstrap';

export default class TextMessage extends Component {
    width = Math.min(80, 5 + this.props.text.length * 1.8);
    style = {
        text: {
            height: '10%',
            width: this.width + '%',
            fontSize: '80%',
            resize: "none",
            color: this.props.color,
            fontWeight: "bold",
        },
    };

    render() {
        return <FormControl
            componentClass={"textarea"}
            value={this.getText()}
            placeholder={"Type a message..."}
            bsSize={"lg"}
            wrap={"hard"}
            rows={2}
            readOnly={"readOnly"}
            style={this.style.text}
        />
    }

    getText = () => {
        return `${this.props.origin}:\n${this.props.text}`;
    };
}