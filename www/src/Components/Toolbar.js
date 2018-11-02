import React, {Component} from 'react';
import {Button, Col, ControlLabel, Form, FormControl, FormGroup, HelpBlock, Label, Row} from 'react-bootstrap';

export default class Toolbar extends Component {
    constructor(props) {
        super(props);

        this.state = {
            labelKey: "",
            metakey: "",
            filename: "",
        };
    }

    styles = {
        toolbar: {
            padding: "16px"
        },
        inputfile: {
            width: "0.1px",
            height: "0.1px",
            opacity: 0,
            overflow: "hidden",
            position: "absolute",
            zIndex: -1,
        },
        shareButton: {
            width: '100%',
            color: "dodgerblue",
            backgroundColor: "#f0f0f0",
            textAlign: "center",
            fontSize: '120%',
            fontWeight: 'bold',
            cursor: "pointer",
            borderRadius: "6px",
        },
        label: {
            backgroundColor: "dodgerblue",
        },
        text: {
            height: '70%',
            width: '100%',
            color: 'MidnightBlue',
            textAlign: "center",
            fontSize: '80%',
            resize: "none",
        },
        downloadButton: {
            height: '25%',
            width: '100%',
            color: 'dodgerblue',
            fontSize: '110%',
            fontWeight: 'bold',
        }
    };

    render() {
        return (
            <Row style={this.styles.toolbar}>
                <Col md={2}>
                    <Form>
                        <FormGroup controlId={"share"}>
                            <FormControl
                                style={this.styles.inputfile}
                                type={"file"}
                                onChange={this.fileUploaded}
                            />
                            <ControlLabel style={this.styles.shareButton}>
                                Share File...
                            </ControlLabel>
                        </FormGroup>
                    </Form>
                    <Label style={this.styles.label}>
                        {this.state.labelKey}
                    </Label>
                    <Form onSubmit={this.download}>
                        <FormGroup
                            controlId={"download"}
                            validationState={this.validationState()}
                        >
                            <FormControl
                                type={"text"}
                                value={this.state.metakey}
                                placeholder={"metakey"}
                                onChange={this.metakeyChange}
                                bsSize={"sm"}
                                style={this.styles.text}
                            />
                            <FormControl
                                type={"text"}
                                value={this.state.filename}
                                placeholder={"filename"}
                                onChange={this.filenameChange}
                                bsSize={"sm"}
                                style={this.styles.text}
                            />
                            <Button type={"submit"} style={this.styles.downloadButton}>
                                Download...
                            </Button>
                            <HelpBlock style={{fontWeight: "bold"}}>
                                {this.state.help}
                            </HelpBlock>
                        </FormGroup>
                    </Form>
                </Col>
            </Row>
        )
    }

    fileUploaded = async (event) => {
        const file = event.target.files[0];
        this.props.shareFile(file, metakey => {
            this.setState({labelKey: metakey});
            setTimeout(() => this.setState({labelKey: ""}), 5000);
        });
    };

    validationState = () => {
        if (this.state.metakey.length === 0) {
            return null;
        }
        if (this.state.metakey.length === 44 &&
            this.state.metakey[this.state.metakey.length - 1] === '=' &&
            this.state.filename.length > 0) {
            return "success";
        }
        if (this.state.metakey.length < 44) {
            return "warning";
        }
        return "error";
    };

    download = (event) => {
        event.preventDefault();
        if (this.validationState() === "success") {
            this.props.download(
                this.state.metakey,
                this.state.filename,
                (success, error) => {
                    if (!success) {
                    console.log({success: success, error: error});
                        this.setState({labelKey: "error " + error});
                        setTimeout(() => this.setState({labelKey: ""}), 3000);
                    }
                }
            );
            this.setState({metakey: "", filename: "", help: ""});
        } else if (this.state.filename.length === 0) {
            this.setState({help: "please type a filename"})
        } else {
            this.setState({help: "wrong metakey format"})
        }
    };

    metakeyChange = (event) => {
        this.setState({metakey: event.target.value});
    };

    filenameChange = (event) => {
        this.setState({filename: event.target.value})
    };
}