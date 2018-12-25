import React, {Component} from 'react';


class Move extends Component {
    constructor(props) {
        super(props);
        this.size = (this.props.size) ? this.props.size : 256;
        this.state = {
            highlighted: false,
        };
    }

    render() {
        return (
            <div style={styles.moveBox}>
                <div
                    style={styles.imageBox(this.size)}
                    onMouseEnter={this.mouseEnter}
                    onMouseLeave={this.mouseLeave}
                    onClick={this.toggleSelected}
                >
                    <img style={{width: '100%'}} src={this.getImageSource()} alt={this.props.move}/>
                </div>
            </div>
        );
    }

    getImageSource = () => {
        const suffix = (this.props.selected || this.state.highlighted) ?
            "" : "_off";
        return `/${this.props.move}${suffix}.png`;
    };

    mouseEnter = () => this.setState({highlighted: true});
    mouseLeave = () => this.setState({highlighted: false});
    toggleSelected = () => this.props.onClick();
}

export default Move;

const styles = {
    moveBox: {
        display: 'flex',
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
    },
    imageBox: (size) => ({
        width: `${size}px`,
        cursor: 'pointer',
    }),
};