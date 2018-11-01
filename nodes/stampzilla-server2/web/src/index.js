import 'bootstrap/dist/js/bootstrap.bundle';
import 'admin-lte/dist/js/adminlte';
import { Provider } from 'react-redux';
import { BrowserRouter as Router } from 'react-router-dom';
import { render } from 'react-dom';
import React from 'react';

import './index.scss';
import App from './containers/App';
import Routes from './routes';
import Websocket from './containers/Websocket';
import store from './store';

render(
  <Provider store={store}>
    <App>
      <Websocket url={`${window.location.protocol.match(/^https/) ? 'wss' : 'ws'}://${window.location.host}/ws`} />

      <Router>
        <Routes />
      </Router>
    </App>
  </Provider>,
  document.getElementById('app'),
);
