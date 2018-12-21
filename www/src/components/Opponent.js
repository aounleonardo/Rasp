import React, {Component} from 'react';
import colors from "./colors";

class Opponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            highlighted: false,
        };
    }

    render() {
        return (
            <div style={styles.opponentBox}>
                <div
                    style={styles.opponentButton(this.isSelected())}
                    onMouseEnter={this.mouseEnter}
                    onMouseLeave={this.mouseLeave}
                    onClick={this.toggleSelected}
                >
                    {this.getButtonContent()}
                </div>
            </div>
        );
    }

    getButtonContent = () => {
        if (this.props.name === "open") {
            return <span role="img" aria-label="Open">Open 📢</span>

        }
        return this.props.name;
    };

    isSelected = () => this.props.selected || this.state.highlighted;

    mouseEnter = () => this.setState({highlighted: true});
    mouseLeave = () => this.setState({highlighted: false});
    toggleSelected = () => this.props.onClick();
}

export default Opponent;

const styles = {
    opponentBox: {
        display: 'flex',
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 20,
    },
    opponentButton: (clicked) => ({
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: 50,
        width: 100,
        backgroundColor: colors.lightBlue,
        color: colors.blue,
        fontWeight: 'bold',
        fontSize: 18,
        borderStyle: (clicked) ? 'solid' : 'double',
        borderRadius: 12,
        borderWidth: 4,
        borderColor: colors.blue,
    }),
};