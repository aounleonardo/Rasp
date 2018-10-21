import React, {Component} from 'react';
import {Button, Col, Label} from 'react-bootstrap';

export default class Chats extends Component {
    constructor(props) {
        super(props);

        this.createRow = this.createRow.bind(this);
        this.onChatSelection = this.onChatSelection.bind(this);
    }

    render() {
        return <Col>
            <h4>
                <Label bsStyle={"primary"}>Chats</Label>
            </h4>
            <h4>
                <Button
                    bsStyle={(this.props.current === "")? "primary" : "info"}
                    onClick={() => this.onChatSelection("")}
                >
                    Home 🏠
                </Button>
            </h4>
            {this.props.peers.sort().map(
                (peer) => this.createRow(peer),
            )}
        </Col>

    }

    createRow = (peer) => {
        return <h6 key={peer}>
            <Button
                bsStyle={(this.props.current === peer) ? "primary" : "info"}
                onClick={() => this.onChatSelection(peer)}
            >
                {peer}
            </Button>
        </h6>
    };

    onChatSelection = (peer) => {
        this.props.chatSelected(peer);
    }
}