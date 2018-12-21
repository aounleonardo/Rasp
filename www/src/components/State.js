import React, {Component} from 'react';
import colors from "./colors";

export default class State extends Component {
    constructor(props) {
        super(props);
    }

    render() {
        return (
            <div style={styles.state}>
                YOYOYO
            </div>
        )
    }
}

const styles = {
    state: {
        width: '100%',
        backgroundColor: colors.lightBlue,
    },
};