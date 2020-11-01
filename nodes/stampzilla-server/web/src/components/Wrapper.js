import { BrowserRouter as Router } from 'react-router-dom';
import { connect } from 'react-redux';
import React from 'react';

import Landing from './Landing';
import Routes from '../routes';
import Websocket from './Websocket';

const Wrapper = (props) => {
  const { server, connection } = props;

  const secure = (window.location.protocol.match(/^https/) || server.get('secure'))
    && connection.get('code') !== 4001;

  return (
    <>
      <Websocket />
      {!secure && <Landing />}
      {secure && (
        <Router>
          <Routes />
        </Router>
      )}
    </>
  );
};

const mapToProps = (state) => ({
  server: state.get('server'),
  connection: state.get('connection'),
});

export default connect(mapToProps)(Wrapper);
