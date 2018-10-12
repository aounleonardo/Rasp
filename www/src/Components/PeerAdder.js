import React, {Component} from 'react';
import {
    Button,
    Col,
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
            height: '100%',
            width: '90%',
            backgroundColor: "red",
        },
        group: {
            padding: '2vmin',
            height: '100%',
            width: '100%',
            backgroundColor: "yellow",
        },
        text: {
            height: '70%',
            width: '100%',
            maxLength: 5,
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
            <Col>
                <Form inline onSubmit={this.add} style={this.style.form}>
                    <FormGroup
                        controlId={"addPeer"}
                        validationState={this.validationState()}
                    >
                        <Row>
                            <FormControl
                                type={"text"}
                                value={this.state.address}
                                placeholder={"address"}
                                onChange={this.addressChange}
                                bsSize={"sm"}
                            />
                        </Row>
                        <Row>
                            <FormControl
                                type={"text"}
                                value={this.state.port}
                                placeholder={"port"}
                                onChange={this.portChange}
                                bsSize={"sm"}
                            />
                        </Row>
                        <Row>
                            <Button type={"submit"}>Add peer</Button>
                        </Row>
                        <HelpBlock>{this.state.help}</HelpBlock>
                    </FormGroup>
                </Form>
            </Col>
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