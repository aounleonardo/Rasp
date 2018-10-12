import React, {Component} from 'react';
import {
    Button,
    Form,
    FormControl,
    FormGroup,
    Row,
    HelpBlock,
} from 'react-bootstrap';

export default class PeerAdder extends Component {
    constructor(props) {
        super(props);

        this.add = this.add.bind(this);
        this.addressChange = this.addressChange.bind(this);
        this.portChange = this.portChange.bind(this);
        this.getPort = this.getPort.bind(this);
        this.getAddress = this.getAddress.bind(this);

        this.state = {
            address: "",
            port: "",
            help: "",
        };
    }

    shouldComponentUpdate(nextProps, nextState) {
        return !(this.state.help !== nextState.help);
    }

    style = {
        form: {
            height: '90%',
            width: '90%',
        },
        group: {
            paddingTop: '2vmin',
            paddingLeft: '3vmin',
            height: '100%',
            width: '100%',
        },
        text: {
            height: '70%',
            width: '60%',
            color: 'MidnightBlue',
            textAlign: "center",
            fontSize: '130%',
            resize: "none",
        },
        button: {
            height:'25%',
            width: '60%',
            color: 'dodgerblue',
            fontSize: '120%',
            fontWeight: 'bold',
        }
    };

    render() {
        return (
                <Form inline onSubmit={this.add} style={this.style.form}>
                    <FormGroup
                        controlId={"addPeer"}
                        validationState={this.validationState()}
                        style={this.style.group}
                    >
                        <Row>
                            <FormControl
                                type={"text"}
                                value={this.state.address}
                                placeholder={"address"}
                                onChange={this.addressChange}
                                bsSize={"sm"}
                                style={this.style.text}
                            />
                        </Row>
                        <Row>
                            <FormControl
                                type={"text"}
                                value={this.state.port}
                                placeholder={"port"}
                                onChange={this.portChange}
                                bsSize={"sm"}
                                style={this.style.text}
                            />
                        </Row>
                        <Row>
                            <Button type={"submit"} style={this.style.button}>Add peer</Button>
                        </Row>
                        <HelpBlock>{this.state.help}</HelpBlock>
                    </FormGroup>
                </Form>
        )
    }

    addressChange = (event) => {
        this.setState({address: event.target.value});
    };

    portChange = (event) => {
        this.setState({port: event.target.value});
    };

    validationState = () => {
        const port = this.getPort();
        if (isNaN(port)) {
            return null;
        } else if (port < 1024) {
            return "warning";
        } else {
            if (this.state.address.length === 0) {
                return null;
            }
            const address = this.getAddress();
            if (address.length !== 4) {
                return "warning";
            }
            return (address.every((byte) => byte >= 0 && byte < 256)) ?
                "success" :
                "error";
        }
    };

    getPort = () => {
        return parseInt(this.state.port);
    };

    getAddress = () => {
        return this.state.address.split('.').map((str) => {
            const parsed = parseInt(str);
            return isNaN(parsed) ? 0 : parsed;
        });
    };

    add = (event) => {
        event.preventDefault();
        if(this.validationState() === "success") {
            this.props.onAdd(
                this.getAddress().join('.'),
                this.getPort().toString(),
            );
            this.setState({address: "", port: "", help: ""})
        } else {
            this.setState({help: "Bad address:port"});
        }
    }
}