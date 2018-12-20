import React, {Component} from 'react';
import colors from "./colors";
import Arena from "./Arena";
import State from "./State";

const db = {
    peers: [
        "Lucas",
        "Remi",
    ],
};

export default class Rasp extends Component {
    constructor(props) {
        super(props);
        this.state = {
            port: this.getServerPort(),
            peers: db.peers,
        };
    }

    render() {
        return (
            <div style={styles.rasp}>
                <Arena
                    style={styles.arena}
                />
                <State
                    style={styles.state}
                />
            </div>
        );
    }

    getServerPort = () => {
        let port = this.props.match.params.serverPort;
        if (!port) {
            port = 8080;
        }
        return port;
    }
}

const styles = {
    rasp: {
        display: 'flex',
        height: '100%',
    },
    arena: {
        display: 'flex',
        flex: 5,
        height: '100%',
        backgroundColor: colors.lightBlue,
    },
    state: {
        display: 'flex',
        flex: 2,
        height: '100%',
        backgroundColor: colors.beige,
    },
};