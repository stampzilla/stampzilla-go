import { BrowserRouter as Router } from 'react-router-dom';
import { connect } from 'react-redux';
import React from 'react';

import App from './App';
import Landing from './Landing';
import Routes from '../routes';
import Websocket from './Websocket';

const Wrapper = (props) => {
  const { server, connection } = props;

  const secure = (window.location.protocol.match(/^https/) || server.get('secure'))
    && connection.get('code') !== 4001;

  return (
    <React.Fragment>
      <Websocket />
      {!secure && <Landing />}
      {secure && (
        <Router>
          <App>
            <Routes />
          </App>
        </Router>
      )}
    </React.Fragment>
  );
};

const mapToProps = state => ({
  server: state.get('server'),
  connection: state.get('connection'),
});

export default connect(mapToProps)(Wrapper);
