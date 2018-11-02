import React, {Component} from 'react';
import {Button, Form, FormControl, FormGroup} from 'react-bootstrap';

export default class Chatbox extends Component {
    constructor(props) {
        super(props);

        this.state = {
            value: '',
        };
    }

    style = {
        box: {
            height: "calc(20vh + 20px)",
            backgroundColor: "dodgerblue",
        },
        form: {
            height: '100%',
            width: '100%',
        },
        group: {
            padding: '2vmin',
            height: '100%',
            width: '100%',
        },
        text: {
            height: '70%',
            width: '100%',
            color: 'MidnightBlue',
            fontSize: '150%',
            resize: "none",
        },
        button: {
            height:'25%',
            width: '100%',
            color: 'dodgerblue',
            fontSize: '150%',
            fontWeight: 'bold',
        }
    };

    render() {

        return (
            <Form inline onSubmit={this.send} style={this.style.form}>
                <FormGroup
                    controlId={"chatText"}
                    validationState={this.validationState()}
                    style={this.style.group}
                >
                    <FormControl
                        componentClass={"textarea"}
                        value={this.state.value}
                        placeholder={"Type a message..."}
                        onChange={this.textChange}
                        bsSize={"lg"}
                        wrap={"hard"}
                        style={this.style.text}
                    />
                    <Button
                        type={"submit"}
                        style={this.style.button}
                    >
                        Send
                    </Button>
                </FormGroup>
            </Form>
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
        } else {
            console.log('incomplete');
        }
        this.setState({value: ''});
    };
}