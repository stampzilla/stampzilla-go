import 'bootstrap/dist/js/bootstrap.bundle'
import 'admin-lte/dist/js/adminlte';

import './index.scss';

import React from "react";
import { Provider } from 'react-redux';
import { BrowserRouter as Router } from 'react-router-dom';
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
