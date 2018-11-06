import React, { Component } from 'react';
import { connect } from 'react-redux';
import { get } from 'axios';
import Url from 'url';
import classnames from 'classnames';

import { update } from '../ducks/app';
import SocketModal from '../components/SocketModal';

class Landing extends Component {
  state = {
    socketModal: false,
  }

  componentWillReceiveProps(props) {
    if (props.server.get('tlsPort') !== this.props.server.get('tlsPort')) {
      console.log('got new TLS port', props.server.get('tlsPort'));

      get(`https://localhost:${props.server.get('tlsPort')}/`).then(() => {
        console.log('SUCCESS');
      }).catch((err) => {
        console.log('FAIL', err);
      });
    }
  }

  onGoSecureClick = () => () => {
    const { dispatch, server } = this.props;

    const url = Url.format({
      protocol: 'wss:',
      hostname: 'localhost',
      port: server.get('tlsPort'),
      pathname: '/ws'
    });
    dispatch(update({ url: url }));
  }

  render = () => {
    const { connected, dispatch, server } = this.props;
    const { socketModal } = this.state;

    return (
      <React.Fragment>
        <SocketModal
          visible={socketModal}
          onClose={() => this.setState({ socketModal: false })}
          onChange={() => dispatch(update({ url: 'ws://localhost:8080/ws' }))}
        />
        {connected === false &&
        <div className="p-4 bg-danger" >
          Not connected!

          <button
            className="btn btn-secondary float-right"
            style={{ marginTop: '-8px' }}
            onClick={() => this.setState({ socketModal: true })}
          >
            Change socket url
          </button>

        </div>
        }
        <div className="d-flex flex-column justify-content-center  align-items-center">

          <h1>stampzilla-go</h1>
          <h2>{server.get('name') || '-'}</h2>
          <a
            href={`http://localhost:${server.get('port')}/ca.crt`}
            className={classnames({
              ["btn btn-outline-secondary mt-3"]: true,
              "disabled": !server.get('port'),
            })}
          >
            Download CA certificate
          </a>

          <button
            className="btn btn-primary mt-4"
            disabled={!server.get('tlsPort')}
            onClick={this.onGoSecureClick()}
          >Go secure
          </button>
        </div>
      </React.Fragment>
    );
  };
}

const mapToProps = state => ({
  connected: state.getIn(['connection', 'connected']),
  server: state.getIn(['server']),
});

export default connect(mapToProps)(Landing);
