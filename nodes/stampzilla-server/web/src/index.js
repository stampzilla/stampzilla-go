import 'bootstrap/dist/js/bootstrap.bundle';
import 'admin-lte/dist/js/adminlte';
import { Provider } from 'react-redux';
import { render } from 'react-dom';
import React from 'react';

import './index.scss';
import ErrorBoundary from './components/ErrorBoundary';
import Wrapper from './components/Wrapper';
import store from './store';

import './images/android-chrome-192x192.png';
import './images/android-chrome-512x512.png';
import './images/apple-touch-icon.png';
import './images/browserconfig.xml';
import './images/favicon-16x16.png';
import './images/favicon-32x32.png';
import './images/favicon.ico';
import './images/safari-pinned-tab.svg';
import './images/site.webmanifest';

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
        .then(() => console.log("Service Worker registered successfully.")) // eslint-disable-line
        .catch(error => console.log('Service Worker registration failed:', error),
        ); // eslint-disable-line
    }
  }());
}
