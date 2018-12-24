import React, {Component} from 'react';
import Arena from "./Arena";
import State from "./State";
import {raspRequest} from "../utils/requests";

const challenges = {
    "Matches": {
        "0000": {
            "Identifier": "0000",
            "Attacker": "A",
            "Defender": "B",
            "Bet": 20,
            "AttackMove": 0,
            "DefenseMove": null,
            "Nonce": 0,
            "HiddenMove": "a",
            "Stage": 0,
        },
        "0001": {
            "Identifier": "0001",
            "Attacker": "A",
            "Defender": "C",
            "Bet": 10,
            "AttackMove": 2,
            "DefenseMove": null,
            "Nonce": 0,
            "HiddenMove": "a",
            "Stage": 0,
        },
        "1000": {
            "Identifier": "1000",
            "Attacker": "C",
            "Defender": "A",
            "Bet": 10,
            "AttackMove": null,
            "DefenseMove": null,
            "Nonce": 0,
            "HiddenMove": "a",
            "Stage": 0,
        },
        "1111": {
            "Identifier": "1111",
            "Attacker": "A",
            "Defender": "B",
            "Bet": 20,
            "AttackMove": 1,
            "DefenseMove": 0,
            "Nonce": 0,
            "HiddenMove": "a",
            "Stage": 2,
        },
    },
    "Proposed": ["0000", "0001"],
    "Pending" : ["1000"],
    "Accepted" : [],
    "Ongoing" : ["1111"],
    "Finished" : [],
};

export default class Rasp extends Component {
    constructor(props) {
        super(props);
        this.endPoint = `http://127.0.0.1:${this.getGossiperPort()}/`;
        this.state = {
            name: "",
            loading: true,
            players: {},
        };
        this.getGossiperName().finally();
        setInterval(this.getPlayers(), 1000);
    }

    static renderLoading() {
        return (
            <div style={styles.loadingContainer}>
                <img
                    style={styles.loadingImage}
                    src={'/logo.svg'}
                    alt={"loading"}
                />
            </div>
        );
    }

    render() {
        return (this.state.loading) ?
            Rasp.renderLoading() :
            <div style={styles.rasp}>
                <div style={styles.arena}>
                    <Arena
                        playerName={this.state.name}
                        players={this.state.players}
                        gossiper={this.endPoint}
                    />
                </div>
                <div style={styles.state}>
                    <State
                        name={this.state.name}
                        balance={this.state.players[this.state.name]}
                        challenges={challenges}
                    />
                </div>
            </div>
    }

    getGossiperPort = () => {
        let port = this.props.match.params.gossiperPort;
        if (!port) {
            port = 8000;
        }
        return port;
    };

    getGossiperName = async () => {
        await raspRequest(
            this.endPoint,
            'identifier/',
            null,
            (name) => {
                this.setState({name: name, loading: false});
            }
        )
    };

    getPlayers = () => {
        raspRequest(
            this.endPoint,
            'players/',
            null,
            (players) => {
                if (players != null && "Players" in players) {
                    this.setState({players: players.Players});
                }
            },
        )
            .finally();
    };
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
    loadingContainer: {
        display: 'flex',
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        height: '100%',
    },
    loadingImage: {
        width: '50%',
    }
};