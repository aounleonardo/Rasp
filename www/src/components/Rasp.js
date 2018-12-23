import React, {Component} from 'react';
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
            gossiper: this.getGossiperPort(),
            peers: db.peers,
        };
    }

    render() {
        return (
            <div style={styles.rasp}>
                <div style={styles.arena}>
                    <Arena
                        opponents={db.peers}
                        gossiper={`http://127.0.0.1:${this.state.gossiper}/`}
                    />
                </div>
                <div style={styles.state}>
                    <State/>
                </div>
            </div>
        );
    }

    getGossiperPort = () => {
        let port = this.props.match.params.gossiperPort;
        if (!port) {
            port = 8000;
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