import React, {Component} from 'react';
import './App.css';
import { BrowserRouter as Router, Route } from 'react-router-dom';
import Rasp from "./components/Rasp";

class App extends Component {
    render() {
        return (
            <Router>
                <div style={{height: '100%'}}>
                    <Route path={"/"} component={Rasp}/>
                    <Route path={"/:serverPort"} component={Rasp}/>
                </div>
            </Router>
        );
    }
}

export default App;
