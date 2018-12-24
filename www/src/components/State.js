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
                    {challengeStates.map((challengeState) => (
                        <Set
                            key={challengeState}
                            name={challengeState}
                            challenges={this.getChallengesForState(
                                this.props.challenges[challengeState],
                                this.props.challenges.Matches,
                            )}
                            clickCallback={() => {
                            }}
                        />
                    ))}
                </div>
            </div>
        )
    }

    getChallengesForState = (stateChallenges, matches) => {
        let challenges = {};
        for (const challenge of stateChallenges) {
            challenges[challenge] = matches[challenge];
        }
        console.log(challenges);
        return challenges;
    }
}

const Set = ({name, challenges}) => {
    return (
        <div>
            <div>
                {`${name}`}
            </div>
            {Object.keys(challenges).map((id, index) =>
                (
                    <Challenge
                        key={`${name}:${index}`}
                        type={name}
                        challenge={challenges[id]}
                        primaryColor={index % 2 === 0}
                    />
                )
            )}
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