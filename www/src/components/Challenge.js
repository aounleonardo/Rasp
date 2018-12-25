import React, {Component} from 'react';
import colors from "./colors";
import {moves, stages, Stages} from "../utils/rasp";
import Move from "./Move";

const done = {
    ACCEPTED: "ACCEPTED",
    CANCELLED: "CANCELLED",
};

class Challenge extends Component {
    constructor(props) {
        super(props);
        this.state = {
            highlighted: false,
            clicked: false,
            done: false,
        }
    }

    renderDone = () =>
        <div style={styles.action(styles.clicked)}>{this.state.done}</div>;


    renderClicked = () => {
        switch (this.props.type) {
            case "Proposed":
                return <this.Cancelled/>;
            case "Pending":
                return <this.Accepted/>;
            case "Ongoing":
                if (this.props.challenge.Stage === Stages.ATTACK) {
                    return <this.Cancelled/>;
                }
                break;
            default:
                break;
        }
    };

    renderContent = () => {
        if (this.state.done) {
            return this.renderDone();
        }
        if (this.state.clicked) {
            return this.renderClicked()
        }
        if (this.state.highlighted) {
            switch (this.props.type) {
                case "Proposed":
                    return <this.Cancel/>;
                case "Pending":
                    return <this.Accept/>;
                case "Ongoing":
                    if (this.props.challenge.Stage === Stages.ATTACK) {
                        return <this.Cancel/>;
                    }
                    break;
                default:
                    break;
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

    actionClicked = () => this.setState({clicked: true});

    Cancel = () => (
        <div
            style={styles.action(styles.cancel)}
            onClick={this.actionClicked}
        >
            CANCEL
        </div>
    );

    Accept = () => (
        <div
            style={styles.action(styles.accept)}
            onClick={this.actionClicked}
        >
            ACCEPT
        </div>
    );

    Clicked = (Buttons) => () => (
        <div style={styles.action(styles.clicked)}>
            <div style={{display: 'flex', flex: 1}}/>
            <Buttons/>
            <div style={{display: 'flex', flex: 1}}/>
        </div>
    );

    Cancelled = this.Clicked(() => (
        <div style={styles.buttons}>
            <Move
                key={"cancel-no"}
                move={"no"}
                size={50}
                selected={false}
                onClick={this.resetState}
            />
            <Move
                key={"cancel-yes"}
                move={"yes"}
                size={50}
                selected={false}
                onClick={() => this.setState({done: done.CANCELLED})}
            />
        </div>
    ));

    Accepted = this.Clicked(() => (
        <div style={styles.buttons}>
            <Move
                key={"accept-no"}
                move={"no"}
                size={50}
                selected={false}
                onClick={this.resetState}
            />
            <div style={{display: 'flex', flex: .5}}/>
            {Object.values(moves).map((move) => (
                <Move
                    key={`challenge-${move}`}
                    move={move}
                    size={50}
                    selected={false}
                    onClick={() => this.setState({done: done.ACCEPTED})}
                />
            ))}
        </div>
    ));

    resetState = () => this.setState({
        highlighted: false,
        clicked: false,
        done: false,
    });
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
    action: (type) => ({
        display: 'flex',
        flex: 1,
        height: '100%',
        justifyContent: 'center',
        alignItems: 'center',
        fontSize: 18,
        fontWeight: 'bold',
        fontFamily: "Helvetica",
        color: colors.white,
        cursor: 'pointer',
        ...type,
    }),
    accept: {
        backgroundColor: 'green',
    },
    cancel: {
        backgroundColor: 'black',
    },
    clicked: {
        backgroundColor: colors.grey,
    },
    buttons: {
        display: 'flex',
        flex: 4,
        height: '100%',
        justifyContent: 'space-around',
        alignItems: 'center',
    },
    buttonContainer: {
        width: 25,
        height: 25,
        backgroundColor: 'red',
    }
};