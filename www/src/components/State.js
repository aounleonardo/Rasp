import React, {Component} from 'react';
import colors from "./colors";
import Challenge from "./Challenge";

const challengeStates = [
    "Proposed",
    "Pending",
    "Accepted",
    "Ongoing",
    "Finished",
];

export default class State extends Component {
    constructor(props) {
        super(props);
    }

    render() {
        return (
            <div style={styles.state}>
                <div style={styles.playerContainer}>
                    <div style={styles.playerLabels}>
                        {this.props.name}
                    </div>
                    <div style={styles.playerLabels}>
                        {this.props.balance}
                    </div>
                </div>
                <div style={styles.scrollContainer}>
                    {challengeStates.map((state) => (
                        <Set
                            key={state}
                            name={state}
                            challenges={[]}
                            clickCallback={() => {}}
                        />
                    ))}
                </div>
            </div>
        )
    }
}

const Set = ({name, challenges, clickCallback}) => {
    return (
        <div>
            <div>
                {`${name}`}
            </div>
            {challenges.forEach((challenge, index) => (
                <Challenge
                    key={`${name}:${index}`}
                    challenge={challenge}
                    primaryColor={index % 2 === 0}
                    onClick={clickCallback}
                />
            ))}
        </div>
    );
};

const styles = {
    state: {
        width: '100%',
        backgroundColor: colors.beige,
    },
    playerContainer: {
        display: 'flex',
        height: 80,
        backgroundColor: colors.lightBlue,
    },
    playerLabels: {
        display: 'flex',
        flex: 1,
        color: colors.blue,
        fontSize: 28,
        fontWeight: 'bold',
        fontFamily: "Helvetica",
        justifyContent: 'center',
        alignItems: 'center',
    },
    scrollContainer: {},
};