import React, {Component} from 'react';
import Move from "./Move";
import Opponent from "./Opponent";
import colors from "./colors";

const moves = ["rock", "paper", "scissors"];

export default class Arena extends Component {
    constructor(props) {
        super(props);
        this.state = {
            selectedMove: "none",
            selectedOpponent: "none",
        }
    }

    render() {
        return (
            <div style={styles.arena}>
                <div style={styles.buttonContainer}>
                    <button style={styles.button}>
                        CHALLENGE
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
                    bet
                </div>
            </div>
        )
    }

    moveSelected = (move) => this.setState({selectedMove: move});
    opponentSelected =
        (opponent) => this.setState({selectedOpponent: opponent});

    listOpponents = () => {
        if (this.props.opponents.length < 1) {
            return (
                <div style={styles.noFriends}>
                    Sorry you have no friends to play with
                    <span role="img" aria-label="sad">😢</span>
                </div>
            )
        }
        const opponents = (this.props.opponents.length === 1) ?
            this.props.opponents :
            ["open", ...this.props.opponents];
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
    }
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
    button: {
        width: 200,
        height: 50,
        backgroundColor: colors.lightBlue,
        color: colors.blue,
        fontWeight: 'bold',
        fontSize: 18,
        borderRadius: 12,
        borderWidth: 4,
        borderColor: colors.blue,
    },
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
    },
    noFriends: {
        fontFamily: 'Helvetica',
        fontSize: 28,
        fontWeight: 'bold',
        color: colors.beige,
    },
};