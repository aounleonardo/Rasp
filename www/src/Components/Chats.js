import React, {Component} from 'react';
import {Button, ButtonGroup, Col, Label} from 'react-bootstrap';

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
                    bsStyle={(this.props.current === "") ? "primary" : "info"}
                    onClick={() => this.onChatSelection("")}
                >
                    Home 🏠
                </Button>
            </h4>
            <ButtonGroup vertical>
                {
                    this.props.peers
                        .filter((peer) => peer !== this.props.identifier)
                        .sort()
                        .map((peer) => this.createRow(peer))
                }
            </ButtonGroup>
        </Col>

    }

    createRow = (peer) => {
        return (
            <Button
                key={peer}
                bsStyle={(this.props.current === peer) ? "primary" : "info"}
                onClick={() => this.onChatSelection(peer)}
            >
                {peer}
            </Button>

        )
    };

    onChatSelection = (peer) => {
        this.props.chatSelected(peer);
    }
}