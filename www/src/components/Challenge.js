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
            DefenceMove,
            Bet,
            Stage,
        } = this.props.challenge;
        return (
            <React.Fragment>
                <div style={styles.infoContainer}>
                    <div style={styles.info}>
                        {Attacker}
                    </div>
                    <Move
                        key={`attackerMove-${Stage}`}
                        move={moves[AttackMove]}
                        size={25}
                        selected={true}
                        onClick={() => {
                        }}
                    />
                </div>
                <div style={styles.infoContainer}>
                    <div style={styles.info}>
                        {Bet}
                    </div>
                    <div style={styles.info}>
                        {stages[Stage]}
                    </div>
                </div>
                <div style={styles.infoContainer}>
                    <div style={styles.info}>
                        {Defender || (
                            <Move
                                key={"defender"}
                                move={null}
                                size={25}
                                selected={true}
                                onClick={() => {
                                }}
                            />
                        )}
                    </div>
                    <Move
                        key={`defenderMove-${Stage}`}
                        move={moves[DefenceMove]}
                        size={25}
                        selected={true}
                        onClick={() => {
                        }}
                    />
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

    onMouseEnter = () => {
        // TODO remove
        console.log(this.props.challenge.Identifier)
        this.setState({highlighted: true})
    };
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
                    onClick={() => this.props.onAccept(
                        this.props.identifier,
                        move,
                    )}
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
    infoContainer: {
        display: 'flex',
        flex: 1,
        flexDirection: 'column',
        height: '100%',
        alignItems: 'center',
    },
    info: {
        display: 'flex',
        flex: 1,
        alignItems: 'center',
        fontSize: 16,
        fontWeight: 'bold',
        fontFamily: "Helvetica",
        color: colors.blue,
        textTransform: 'uppercase',
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