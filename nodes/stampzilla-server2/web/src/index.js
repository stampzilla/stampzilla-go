import 'bootstrap/dist/js/bootstrap.bundle';
import 'admin-lte/dist/js/adminlte';
import { Provider } from 'react-redux';
import { render } from 'react-dom';
import React from 'react';

import './index.scss';
import Wrapper from './components/Wrapper';
import store from './store';

render(
  <Provider store={store}>
    <Wrapper />
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
