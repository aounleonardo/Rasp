import React, {Component} from 'react';
import {Button, Row} from 'react-bootstrap';

export default class PeerAdder extends Component {
    constructor(props) {
        super(props);

        this.add = this.add.bind(this);
    }

    render() {
        return (
            <form onSubmit={this.add}>
                <Row>
                    <input
                        type={"text"}
                        ref={"address"}
                        placeholder={"address"}
                    />
                </Row>
                <Row>
                    <input type={"text"} ref={"port"} placeholder={"port"}/>
                </Row>
                <Row>
                    <Button type={"submit"}>Add peer</Button>
                </Row>
            </form>
        )
    }

    add = (event) => {
        event.preventDefault();
        this.props.onAdd(this.refs.address.value, this.refs.port.value);
        this.refs.address.value = "";
        this.refs.port.value = "";
    }
}