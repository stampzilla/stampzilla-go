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
  };

  componentWillReceiveProps(props) {
    if (props.server.get('tlsPort') !== this.props.server.get('tlsPort')) {
      const { app } = this.props;
      const serverUrl = Url.parse(app.get('url'));
      const url = `https://${serverUrl.hostname}:${props.server.get(
        'tlsPort',
      )}/`;

      const check = () =>
        get(url)
          .then(() => {
            const socketUrl = Url.format({
              protocol: 'wss:',
              hostname: serverUrl.hostname,
              port: props.server.get('tlsPort'),
              pathname: '/ws',
            });
            props.dispatch(update({ url: socketUrl }));
          })
          .catch(() => {});
      clearTimeout(this.checker);
      this.checker = setInterval(check, 1000);
      check();
    }
  }

  onGoSecureClick = () => () => {
    const { dispatch, server, app } = this.props;
    const serverUrl = Url.parse(app.get('url'));

    const url = Url.format({
      protocol: 'wss:',
      hostname: serverUrl.hostname,
      port: server.get('tlsPort'),
      pathname: '/ws',
    });
    dispatch(update({ url }));
  };

  render = () => {
    const {
      connected, dispatch, server, app,
    } = this.props;
    const { socketModal } = this.state;
    const url = Url.parse(app.get('url'));

    return (
      <React.Fragment>
        <SocketModal
          visible={socketModal}
          onClose={() => this.setState({ socketModal: false })}
          onChange={({ hostname, port }) =>
            dispatch(update({ url: `ws://${hostname}:${port}/ws` }))
          }
        />
        {connected === false && (
          <div className="p-4 bg-danger">
            Not connected!
            <button
              className="btn btn-secondary float-right"
              style={{ marginTop: '-8px' }}
              onClick={() => this.setState({ socketModal: true })}
            >
              Change socket url
            </button>
          </div>
        )}
        <div className="d-flex flex-column justify-content-center  align-items-center">
          <h1>stampzilla-go</h1>
          {connected === null && (
            <i className="fa fa-refresh fa-spin fa-3x fa-fw" />
          )}
          {connected && (
            <React.Fragment>
              <h2>{server.get('name') || '-'}</h2>

              <pre>
                Hostname: {url.hostname}
                <br />
                HTTP Port: {server.get('port')}
                <br />
                TLS port: {server.get('tlsPort')}
              </pre>

              <a
                href={`http://${window.location.hostname}:${server.get(
                  'port',
                )}/ca.crt`}
                className={classnames({
                  'btn btn-outline-secondary mt-3': true,
                  disabled: !server.get('port'),
                })}
              >
                Download CA certificate
              </a>

              <button
                className="btn btn-primary mt-4"
                disabled={!server.get('tlsPort')}
                onClick={this.onGoSecureClick()}
              >
                Go secure
              </button>
            </React.Fragment>
          )}
        </div>
      </React.Fragment>
    );
  };
}

const mapToProps = state => ({
  connected: state.getIn(['connection', 'connected']),
  server: state.getIn(['server']),
  app: state.getIn(['app']),
});

export default connect(mapToProps)(Landing);
