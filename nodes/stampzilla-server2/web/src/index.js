import 'bootstrap/dist/js/bootstrap.bundle';
import 'admin-lte/dist/js/adminlte';
import { Provider } from 'react-redux';
import { BrowserRouter as Router } from 'react-router-dom';
import { render } from 'react-dom';
import React from 'react';

import './index.scss';
import App from './containers/App';
import Landing from './containers/Landing';
import Routes from './routes';
import Websocket from './containers/Websocket';
import store from './store';

const secure = false;

render(
  <Provider store={store}>
    <React.Fragment>
      <Websocket />
      {!secure &&
        <Landing />
      }
      {secure &&
      <App>

        <Router>
          <Routes />
        </Router>

      </App>
      }
    </React.Fragment>
  </Provider>,
  document.getElementById('app'),
);

if (NODE_ENV === 'production') { // eslint-disable-line no-undef
  (function registerServiceWorker() {
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.register('service-worker.js', { scope: '/' })
        .then(() => console.log('Service Worker registered successfully.'))
        .catch(error => console.log('Service Worker registration failed:', error));
    }
  }());
}
