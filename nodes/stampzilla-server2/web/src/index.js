import React from "react";
import { render } from 'react-dom';
import { Provider } from 'react-redux';
import { BrowserRouter as Router, Route } from 'react-router-dom';

import store from './store';
import Routes from './routes';
import Nodes from "./containers/Nodes";

render(
    <Provider store={store}>
        <Router>
            <Routes /> 
        </Router>
    </Provider>,
    document.getElementById("app")
)
