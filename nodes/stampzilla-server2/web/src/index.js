import 'bootstrap/dist/js/bootstrap.bundle'
import 'admin-lte/dist/js/adminlte';

import './index.scss';

import React from "react";
import { Provider } from 'react-redux';
import { BrowserRouter as Router } from 'react-router-dom';
import { render } from 'react-dom';

import Routes from './routes';
import App from "./containers/App";
import Websocket from "./containers/Websocket";
import store from './store';

render(
    <Provider store={store}>
      <App>
        <Websocket url={(location.protocol.match(/^https/) ? "wss" : "ws" ).concat("://localhost:5000/ws")}/>

        <Router>
          <Routes /> 
        </Router>
      </App>
    </Provider>,
    document.getElementById("app")
)
