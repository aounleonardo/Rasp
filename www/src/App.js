import React, {Component} from 'react';
import './App.css';
import Peerster from "./Components/Peerster";

require('dotenv').config();

class App extends Component {
    render() {
        return (
            <Peerster/>
        );
    }
}

export default App;
