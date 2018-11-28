import React, {Component} from 'react';
import {Button, Form, FormControl, FormGroup, Row} from 'react-bootstrap';
import SearchesWindow from "./SearchesWindow";

export default class FileSearch extends Component {
    constructor(props) {
        super(props);

        this.state = {
            keywords: [],
        }
    }

    styles = {
        form: {
            width: '50%',
        },
        text: {
            width: '100%',
            color: 'MidnightBlue',
            textAlign: "center",
            fontSize: '80%',
            resize: "none",
        },
        searchButton: {
            height: '25%',
            width: '40%',
            color: 'dodgerblue',
            fontSize: '110%',
            fontWeight: 'bold',
        },
    };

    render() {
        return (
            <Row>
                <Form inline onSubmit={this.onSearch} style={this.styles.form}>
                    <FormGroup controlId={"search"}>
                        <FormControl
                            style={this.styles.text}
                            type={"text"}
                            placeholder={"search for comma seperated keywords"}
                            onChange={this.keywordsChange}
                            bsSize={"sm"}
                        />
                        <Button
                            type={"submit"}
                            style={this.styles.searchButton}
                        >
                            Search...
                        </Button>
                    </FormGroup>
                </Form>
                <SearchesWindow
                    searches={this.props.searches}
                    download={this.props.download}
                    keywords={this.state.keywords}
                />
            </Row>
        )
    }

    keywordsChange = (event) => {
        this.setState({keywords: event.target.value.split(',')});
    };

    onSearch = (event) => {
        event.preventDefault();
        this.props.searchFor(this.state.keywords);
    }
}