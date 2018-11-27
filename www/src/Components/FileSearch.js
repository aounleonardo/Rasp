import React, {Component} from 'react';
import {Row, Button, Form, FormControl, FormGroup} from 'react-bootstrap';
import SearchesWindow from "./SearchesWindow";

export default class FileSearch extends Component {
    constructor(props) {
        super(props);

        this.state = {
            keywords: [],
        }
    }

    styles = {};

    render() {
        return (
            <Row>
                <Form inline onSubmit={this.onSearch}>
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