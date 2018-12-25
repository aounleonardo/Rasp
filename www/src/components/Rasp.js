import React, {Component} from 'react';
import Arena from "./Arena";
import State from "./State";
import {raspRequest} from "../utils/requests";
import Move from "./Move";

export default class Rasp extends Component {
    constructor(props) {
        super(props);
        this.endPoint = `http://127.0.0.1:${this.getGossiperPort()}/`;
        this.state = {
            name: "",
            loading: true,
            players: {},
            challenges: {
                Matches: {},
                Proposed: [],
                Pending: [],
                Accepted: [],
                Ongoing: [],
                Finished: [],
            },
        };
        this.getGossiperName().finally();
        setInterval(this.getPlayers, 1000);
        setInterval(this.getChallenges, 1000);
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
                        balance={this.state.players[this.state.name] ||
                        (
                            <Move
                                key={"balance"}
                                move={null}
                                size={25}
                                selected={true}
                                onClick={() => {
                                }}
                            />
                        )}
                        challenges={this.state.challenges}
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
                if (players && players.Players) {
                    this.setState({players: players.Players});
                }
            },
        )
            .finally();
    };

    getChallenges = () => {
        raspRequest(
            this.endPoint,
            'state/',
            null,
            (challenges) => {
                if (challenges) {
                    this.setState({
                        challenges: {
                            Matches: challenges.Matches || {},
                            Proposed: Object.keys(challenges.Proposed || {}),
                            Pending: Object.keys(challenges.Pending || {}),
                            Accepted: Object.keys(challenges.Accepted || {}),
                            Ongoing: Object.keys(challenges.Ongoing || {}),
                            Finished: Object.keys(challenges.Finished || {}),
                        }
                    })
                }
            }
        )
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