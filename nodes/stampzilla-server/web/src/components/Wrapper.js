import { BrowserRouter as Router } from 'react-router-dom';
import { connect } from 'react-redux';
import React from 'react';

import App from './App';
import Landing from './Landing';
import Routes from '../routes';
import Websocket from './Websocket';

const Wrapper = (props) => {
  const { server } = props;

  const secure = window.location.protocol.match(/^https/) || server.get('secure');

  return (
    <React.Fragment>
      <Websocket />
      {!secure &&
        <Landing />
      }
      {secure &&
      <Router>
        <App>
          <Routes />
        </App>
      </Router>
      }
    </React.Fragment>
  );
};

const mapToProps = state => ({
  server: state.getIn(['server']),
});

export default connect(mapToProps)(Wrapper);
