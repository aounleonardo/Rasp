import React, {Component} from 'react';
import {Col, ControlLabel, Form, FormControl, FormGroup, Row} from 'react-bootstrap';

export default class Toolbar extends Component {
    constructor(props) {
        super(props)
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
        button: {
            width: '80%',
            color: "dodgerblue",
            backgroundColor: "#f0f0f0",
            textAlign: "center",
            fontSize: '120%',
            fontWeight: 'bold',
            cursor: "pointer",
            borderRadius: "6px",
        },
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
                            <ControlLabel style={this.styles.button}>
                                Share File...
                            </ControlLabel>
                        </FormGroup>
                    </Form>
                </Col>
            </Row>
        )
    }

    fileUploaded = (event) => {
        console.log(event.target.files[0])
    }
}