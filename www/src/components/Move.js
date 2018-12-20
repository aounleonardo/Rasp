import React, {Component} from 'react';


class Move extends Component {
    constructor(props) {
        super(props);
        this.state = {
            highlighted: false,
        };
    }

    render() {
        return (
            <div style={styles.moveBox}>
                <div
                    style={styles.imageBox}
                    onMouseEnter={this.mouseEnter}
                    onMouseLeave={this.mouseLeave}
                >
                    <img src={this.getImageSource()}/>
                </div>
            </div>
        );
    }

    getImageSource = () => {
        const suffix = (this.state.highlighted) ? "_alt" : "";
        return `/${this.props.move}${suffix}.png`;
    };

    mouseEnter = () => {
        this.setState({highlighted: true});
    };

    mouseLeave = () => {
        this.setState({highlighted: false});
    };
}

export default Move;

const styles = {
    moveBox: {
        display: 'flex',
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
    },
    imageBox: {
        width: 256,
        height: 256,
    },
};