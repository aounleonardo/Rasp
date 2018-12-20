import React, {Component} from 'react';
import Move from "./Move"

export default class Arena extends Component {
    constructor(props) {
        super(props);
        this.style = {...this.props.style, ...styles.arena};
    }

    render() {
        return (
            <div style={this.style}>
                <div style={styles.movesContainer}>
                    <Move move={"rock"}/>
                    <Move move={"paper"}/>
                    <Move move={"scissors"}/>
                </div>
                <div style={styles.opponentContainer}>
                opponent
                </div>
                <div style={styles.betContainer}>
                    bet
                </div>
            </div>
        )
    }
}

const styles = {
    arena: {
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
    },
    movesContainer: {
        display: 'flex',
        flex: 4,
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
};