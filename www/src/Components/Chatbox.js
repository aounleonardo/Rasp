import React, {Component} from 'react';
import {Button} from 'react-bootstrap';

export default class Chatbox extends Component {
    constructor(props) {
        super(props);

        this.send = this.send.bind(this);
    }

    render() {
        return(
            <form onSubmit={this.send}>
                <input type={"text"} ref={"message"} placeholder={"Type a message..."}/>
                <Button type={"submit"}>Send</Button>
            </form>
        )
    }

    send = (event) => {
        event.preventDefault();
        this.props.onSend(this.refs.message.value);
        this.refs.message.value = "";
    }
}