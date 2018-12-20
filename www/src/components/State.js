import React, {Component} from 'react';

export default class State extends Component {
    constructor(props) {
        super(props);
        this.style = {...this.props.style, ...styles.state};
    }

    render() {
        return (
            <div style={this.style}>
            </div>
        )
    }
}

const styles = {
    state: {
    },
};