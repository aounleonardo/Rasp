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
                <div style={styles.arena}>
                    <Arena opponents={db.peers}/>
                </div>
                <div style={styles.state}>
                    <State/>
                </div>
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
    },
    state: {
        display: 'flex',
        flex: 2,
    },
};