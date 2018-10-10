import React, {Component} from 'react';
import {Button, Col, Form, FormControl, FormGroup} from 'react-bootstrap';

export default class Chatbox extends Component {
    constructor(props) {
        super(props);

        this.send = this.send.bind(this);
        this.textChange = this.textChange.bind(this);

        this.state = {
            value: '',
        };
    }

    render() {
        return (
            <Col>
                <Form inline onSubmit={this.send}>
                    <FormGroup
                        controlId={"chatText"}
                        validationState={this.validationState()}
                    >
                        <FormControl
                            type={"text"}
                            value={this.state.value}
                            placeholder={"Type a message..."}
                            onChange={this.textChange}
                            bsSize={"lg"}
                        />
                        <Button type={"submit"}>Send</Button>
                    </FormGroup>
                </Form>
            </Col>
        )
    }

    validationState = () => {
        if (this.state.value.length > 0) {
            return 'success';
        }
        return null;
    };

    textChange = (event) => {
        this.setState({value: event.target.value});
    };

    send = (event) => {
        event.preventDefault();
        if (this.validationState() === 'success') {
            this.props.onSend(this.state.value);
            console.log(this.state.value);
        } else {
            console.log('incomplete');
        }
        this.setState({value: ''});
    };
}