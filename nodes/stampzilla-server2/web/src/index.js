import React from "react";
import { Provider } from 'react-redux';
import { BrowserRouter as Router, Route } from 'react-router-dom';
import { render } from 'react-dom';

import Nodes from "./containers/Nodes";
import Routes from './routes';
import Websocket from "./containers/Websocket";
import store from './store';

render(
    <Provider store={store}>
      <React.Fragment>
        <Websocket url="ws://localhost:5000/ws"/>
        <Router>
            <Routes /> 
        </Router>
      </React.Fragment>
    </Provider>,
    document.getElementById("app")
)
