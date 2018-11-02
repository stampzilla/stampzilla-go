import React, { Component } from 'react';
import { connect } from 'react-redux';

import { update } from '../ducks/app';
import SocketModal from '../components/SocketModal';

class Landing extends Component {
  state = {
    socketModal: false,
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
        {!connected &&
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
          <h2>{server.get('name')}</h2>
          <button
            className="btn btn-outline-secondary mt-3"
            disabled={!server.get('port')}
          >
            Download CA certificate
          </button>

          <button
            className="btn btn-primary mt-4"
            disabled={!server.get('tlsPort')}
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
