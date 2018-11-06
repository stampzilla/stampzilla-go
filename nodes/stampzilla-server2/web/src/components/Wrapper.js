import { BrowserRouter as Router } from 'react-router-dom';
import { connect } from 'react-redux';
import React from 'react';

import App from '../containers/App';
import Landing from '../containers/Landing';
import Routes from '../routes';
import Websocket from '../containers/Websocket';

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
      <App>

        <Router>
          <Routes />
        </Router>

      </App>
      }
    </React.Fragment>
  );
};

const mapToProps = state => ({
  server: state.getIn(['server']),
});

export default connect(mapToProps)(Wrapper);
