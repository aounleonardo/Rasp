import React, {Component} from 'react';
import {Col, Row} from 'react-bootstrap';
import TextMessage from "./TextMessage";
import {sha256} from 'js-sha256';

export default class MessagesWindow extends Component {
    constructor(props) {
        super(props);

        this.getColor = this.getColor.bind(this);
    }

    styles = {
        messageRow: {
            paddingLeft: "5%",
            paddingTop: "8px",
        },
        text: {
            fontWeight: "bold",
            cornerRadius: "30%",
            wrap: "hard",
        },
        messageColors: [
            "dodgerblue",
            "red",
            "orange",
            "green",
            "brown",
            "pink",
        ]
    };

    render() {
        return (
            <Col>
                {this.props.messages.map((message) => this.createRow(message))}
            </Col>
        )
    };

    createRow = (msg) => {
        return <Row
            key={`${msg["Origin"]}:${msg["ID"]}`}
            style={this.styles.messageRow}
        >
            <TextMessage origin={msg["Origin"]} text={msg["Text"]} color={this.getColor(msg["Origin"])}/>
        </Row>
    };

    getColor = (author) => {
        if(author === this.props.identifier) {
            return this.styles.messageColors[0];
        }
        const index = parseInt(sha256.hex(author), 16)
            % (this.styles.messageColors.length - 1);
        return this.styles.messageColors[index + 1];
    }
}