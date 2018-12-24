import React, {Component} from 'react';
import Move from "./Move";
import Opponent from "./Opponent";
import colors from "./colors";
import "./style.css"
import {raspRequest} from "../utils/requests";

const moves = ["rock", "paper", "scissors"];
const initialState = {
    selectedMove: "none",
    selectedOpponent: "none",
    bet: 0,
    betHighlighted: false,
};

export default class Arena extends Component {
    constructor(props) {
        super(props);
        this.state = {...{}, ...initialState};
    }

    render() {
        const state = this.getButtonState();
        return (
            <div style={styles.arena}>
                <div style={styles.buttonContainer}>
                    <button
                        style={styles.button(state === "send")}
                        onClick={this.buttonPressed}
                    >
                        {
                            {
                                "move": "CHOOSE MOVE",
                                "opponent": "PICK OPPONENT",
                                "bet": "PLACE BET",
                                "send": "SEND CHALLENGE",
                            }[state]
                        }
                    </button>
                </div>
                <div style={styles.movesContainer}>
                    {moves.map((move) => (
                        <Move
                            key={move}
                            move={move}
                            selected={move === this.state.selectedMove}
                            onClick={() => this.moveSelected(move)}
                        />
                    ))}
                </div>
                <div style={styles.opponentContainer}>
                    {this.listOpponents()}
                </div>
                <div style={styles.betContainer}>
                    <div style={styles.betLabel}>
                        BET
                    </div>
                    <div
                        style={styles.sliderContainer}
                    >
                        <input
                            type={"range"}
                            min={"1"}
                            max={"100"}
                            value={this.state.bet}
                            style={styles.betSlider}
                            onChange={this.changeBet}
                            className=
                                {`slider${(this.isBetOff()) ? '-off' : ''}`}
                        />
                    </div>
                    <div style={styles.betValue}>
                        {this.getBet()}
                    </div>
                </div>
            </div>
        )
    }

    getOpponents = () => {
        return Object.keys(this.props.players)
            .filter((name) => name !== this.props.playerName)
    };

    moveSelected = (move) => this.setState({selectedMove: move});
    opponentSelected =
        (opponent) => this.setState({selectedOpponent: opponent});

    listOpponents = () => {
        let opponents = this.getOpponents();
        if (opponents.length < 1) {
            return (
                <div style={styles.noFriends}>
                    Sorry you have no friends to play with
                    <span role="img" aria-label="sad">😢</span>
                </div>
            )
        }
        opponents = (opponents.length === 1) ?
            opponents :
            ["open", ...opponents];
        return opponents.map(
            (opponent) => (
                <Opponent
                    key={opponent}
                    name={opponent}
                    selected=
                        {opponent === this.state.selectedOpponent}
                    onClick={() => this.opponentSelected(opponent)}
                />
            )
        );
    };

    changeBet = (event) => {
        this.setState({bet: event.target.value});
    };

    getBet = () => Math.floor(this.state.bet / 10);

    isBetOff = () => !this.state.betHighlighted && this.getBet() === 0;

    betMouseEnter = () => {
        this.setState({betHighlighted: true});
    };

    betMouseLeave = () => {
        this.setState({betHighlighted: false});
    };

    buttonPressed = async () => {
        if (this.getButtonState() !== "send") {
            await raspRequest(this.props.gossiper, 'identifier/', null, (res) => console.log(res));
            return;
        }
        const payload = {
            Destination: (this.state.selectedOpponent === "open") ?
                null : this.state.selectedOpponent,
            Bet: this.getBet(),
            Move: this.state.selectedMove,
        };
        console.log({createMatch: payload});
        await raspRequest(
            this.props.gossiper,
            'create-match/',
            payload,
            (res) => {
                console.log(res);
            }
        );
        this.resetState();
    };

    getButtonState = () => {
        if (this.state.selectedMove === "none") {
            return "move";
        }
        if (this.state.selectedOpponent === "none") {
            return "opponent";
        }
        if (this.getBet() === 0) {
            return "bet";
        }
        return "send";
    };

    resetState = () => {
        this.setState({...{}, ...initialState});
    };
}

const styles = {
    arena: {
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
        width: '100%',
        backgroundColor: colors.grey,
    },
    buttonContainer: {
        display: 'flex',
        flex: 1,
        flexDirection: 'column',
        justifyContent: 'flex-end',
        alignItems: 'center',
    },
    button: (valid) => ({
        width: 200,
        height: 50,
        backgroundColor: (valid) ? colors.lightBlue : colors.lightBlue,
        color: (valid) ? colors.blue : colors.blue,
        fontWeight: 'bold',
        fontSize: 18,
        borderRadius: 12,
        borderWidth: 4,
        borderColor: (valid) ? colors.blue : colors.grey,
        borderStyle: (valid) ? '' : 'double',
    }),
    movesContainer: {
        display: 'flex',
        flex: 3,
        width: '100%',
        justifyContent: 'center',
        alignItems: 'center',
    },
    opponentContainer: {
        display: 'flex',
        flex: 1,
    },
    betContainer: {
        display: 'flex',
        flex: 1,
        width: "100%",
    },
    noFriends: {
        fontFamily: 'Helvetica',
        fontSize: 28,
        fontWeight: 'bold',
        color: colors.beige,
    },
    sliderContainer: {
        display: 'flex',
        flex: 2,
        width: "50%",
        paddingLeft: 20,
        paddingRight: 20,
        justifyContent: 'center',
        alignItems: 'center',
    },
    betLabel: {
        display: 'flex',
        flex: 1,
        justifyContent: 'flex-end',
        alignItems: 'center',
        fontFamily: 'Helvetica',
        fontSize: 28,
        fontWeight: 'bold',
        color: colors.blue,
    },
    betValue: {
        display: 'flex',
        flex: 1,
        justifyContent: 'flex-start',
        alignItems: 'center',
        fontFamily: 'Helvetica',
        fontSize: 28,
        fontWeight: 'bold',
        color: colors.blue,
    },
    betSlider: {
        WebkitAppearance: 'none',
        height: 25,
        WebkitTransition: '.2s',
        backgroundColor: colors.blue,
        borderRadius: 12,
        cursor: 'pointer',
        width: '100%',
    },
};