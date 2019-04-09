import 'bootstrap/dist/js/bootstrap.bundle';
import 'admin-lte/dist/js/adminlte';
import { Provider } from 'react-redux';
import { render } from 'react-dom';
import React from 'react';

import './index.scss';
import ErrorBoundary from './components/ErrorBoundary';
import Wrapper from './components/Wrapper';
import store from './store';

render(
  <Provider store={store}>
    <ErrorBoundary>
      <Wrapper />
    </ErrorBoundary>
  </Provider>,
  document.getElementById('app'),
);

/* eslint-disable no-undef */
if (NODE_ENV === 'production') {
  (function registerServiceWorker() {
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker
        .register('service-worker.js', { scope: '/' })
        .then(() => console.log('Service Worker registered successfully.')) // eslint-disable-line
        .catch(error => console.log('Service Worker registration failed:', error)); // eslint-disable-line
    }
  }());
}
