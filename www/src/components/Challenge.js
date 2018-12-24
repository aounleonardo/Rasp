import React, {Component} from 'react';
import colors from "./colors";
import {moves, stages} from "../utils/rasp";

class Challenge extends Component {
    constructor(props) {
        super(props);
        this.state = {
            highlighted: false,
        }
    }

    renderContent = () => {
        if (this.state.highlighted) {
            switch (this.props.type) {
                case "Proposed":
                    return (
                        <div>
                            CANCEL
                        </div>
                    );
                case "Pending":
                    return (
                        <div>
                            ACCEPT
                        </div>
                    );
            }
        }
        const {
            Attacker,
            Defender,
            AttackMove,
            DefenseMove,
            Bet,
            Stage,
        } = this.props.challenge;
        return (
            <React.Fragment>
                <div style={styles.attackerContainer}>
                    <div style={styles.attackerName}>
                        {Attacker}
                    </div>
                    <div style={styles.attackerMove}>
                        {
                            (AttackMove === null) ?
                                "?" :
                                moves[AttackMove]
                        }
                    </div>
                </div>
                <div style={styles.neutralContainer}>
                    <div style={styles.bet}>
                        {Bet}
                    </div>
                    <div style={styles.stage}>
                        {stages[Stage]}
                    </div>
                </div>
                <div style={styles.defenderContainer}>
                    <div style={styles.defenderName}>
                        {Defender}
                    </div>
                    <div style={styles.defenderMove}>
                        {
                            (DefenseMove === null) ?
                                "?" :
                                moves[DefenseMove]
                        }
                    </div>
                </div>
            </React.Fragment>

        );
    };

    render() {
        const content = this.renderContent();
        return (
            <div
                style={styles.challenge(this.props.primaryColor)}
                onMouseEnter={this.onMouseEnter}
                onMouseLeave={this.onMouseLeave}
            >
                {content}
            </div>
        );
    }

    onMouseEnter = () => this.setState({highlighted: true});
    onMouseLeave = () => this.setState({highlighted: false});

}

export default Challenge;

const styles = {
    challenge: (primary) => ({
        display: 'flex',
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        height: 80,
        backgroundColor: (primary) ? colors.beige : colors.white,
    }),
    attackerContainer: {
        display: 'flex',
        flex: 1,
        flexDirection: 'column',
        height: '100%',
        alignItems: 'center',
    },
    defenderContainer: {
        display: 'flex',
        flex: 1,
        flexDirection: 'column',
        height: '100%',
        alignItems: 'center',
    },
    neutralContainer: {
        display: 'flex',
        flex: 1,
        flexDirection: 'column',
        height: '100%',
        alignItems: 'center',
    },
    attackerName: {
        display: 'flex',
        flex: 1,
        alignItems: 'center',
    },
    defenderName: {
        display: 'flex',
        flex: 1,
        alignItems: 'center',

    },
    attackerMove: {
        display: 'flex',
        flex: 1,
        alignItems: 'center',

    },
    defenderMove: {
        display: 'flex',
        flex: 1,
        alignItems: 'center',

    },
    bet: {
        display: 'flex',
        flex: 1,
        alignItems: 'center',

    },
    stage: {
        display: 'flex',
        flex: 1,
        alignItems: 'center',

    },
};