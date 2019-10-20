import React, { Component } from 'react';
import { connect } from 'react-redux';
import { get, post } from 'axios';
import Url from 'url';
import classnames from 'classnames';

import { update } from '../ducks/app';
import SocketModal from './SocketModal';
import Login from './Login';

class Landing extends Component {
  state = {
    socketModal: false,
    tlsOk: false,
  };

  componentWillReceiveProps(props) {
    if (props.server.get('tlsPort') !== this.props.server.get('tlsPort')) {
      const { app } = this.props;
      const serverUrl = Url.parse(app.get('url'));
      const url = `https://${serverUrl.hostname}:${props.server.get(
        'tlsPort',
      )}/`;

      const check = () => {
        this.setState({ tlsOk: false });
        return get(url)
          .then(() => {
            this.setState({ tlsOk: true });
            clearTimeout(this.checker);
            // const socketUrl = Url.format({
            // protocol: 'wss:',
            // hostname: serverUrl.hostname,
            // port: props.server.get('tlsPort'),
            // pathname: '/ws',
            // });
            // props.dispatch(update({ url: socketUrl }));
          })
          .catch(() => {
            this.setState({ tlsOk: false });
          });
      };
      clearTimeout(this.checker);
      this.checker = setInterval(check, 1000);
      check();
    }
  }

  onGoSecureClick = () => () => {
    const { server, app } = this.props;
    const serverUrl = Url.parse(app.get('url'));

    const url = Url.format({
      protocol: 'wss',
      hostname: serverUrl.hostname,
      port: server.get('tlsPort'),
      pathname: '',
    });

    window.location.href = url;
  };

  render = () => {
    const {
      connected,
      connecting,
      dispatch,
      server,
      app,
      disconnectionCode,
      port,
    } = this.props;
    const { socketModal, tlsOk } = this.state;
    const url = Url.parse(app.get('url'));

    return (
      <div className="landing-container">
        <SocketModal
          visible={socketModal}
          onClose={() => this.setState({ socketModal: false })}
          onChange={({ hostname, p }) => dispatch(update({ url: `ws://${hostname}:${p}/ws` }))
          }
        />
        <div className="background">
          <div className="content d-flex flex-column justify-content-center align-items-center">
            <h1>stampzilla-go</h1>
            {connected !== true && (
              <div>
                {connecting && (
                  <React.Fragment>
                    Trying to connect to
                    {' '}
                    {app.get('url')}
                    ...
                    <i className="fa fa-refresh fa-spin fa-fw" />
                  </React.Fragment>
                )}
                <button
                  type="button"
                  className="btn btn-secondary float-right ml-2"
                  style={{ marginTop: '-8px' }}
                  onClick={() => this.setState({ socketModal: true })}
                >
                  Change socket url
                </button>
              </div>
            )}

            {server.get('port') && (
              <React.Fragment>
                <h2>{server.get('name') || '-'}</h2>

                <pre>
                  Hostname:
                  {' '}
                  {url.hostname}
                  <br />
                  HTTP Port:
                  {' '}
                  {server.get('port')}
                  {server.get('port') === port && connected && (
                    <span className="text-success">
                      {' '}
                      <i className="fa fa-check" />
                      {' '}
(connected)
                    </span>
                  )}
                  <br />
                  TLS port:
                  {' '}
                  {server.get('tlsPort')}
                  {server.get('tlsPort') === port && connected && (
                    <span className="text-success">
                      {' '}
                      <i className="fa fa-check" />
                      {' '}
(connected)
                    </span>
                  )}
                  {!connected
                    && tlsOk
                    && server.get('tlsPort') === port
                    && disconnectionCode === 4001 && (
                      <span className="text-danger">
                        {' '}
                        <i className="fa fa-times" />
                        {' '}
(unauthorized)
                      </span>
                  )}
                </pre>
              </React.Fragment>
            )}

            {!connected
              && server.get('tlsPort') === port
              && disconnectionCode === 4001 && (
                <Login
                  onSubmit={(username, password) => {
                    const serverUrl = Url.parse(app.get('url'));
                    const u = `https://${serverUrl.hostname}:${server.get(
                      'tlsPort',
                    )}/login`;

                    const formData = new FormData();
                    formData.append('username', username);
                    formData.append('password', password);
                    post(u, formData, { withCredentials: true })
                      .then((response) => {
                        this.props.dispatch(update({ url: '' }));
                        this.props.dispatch(update({ url: app.get('url') }));
                      })
                      .catch((error) => {
                        console.log('error', error);
                      });
                  }}
                />
            )}

            {connected && (
              <React.Fragment>
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
                  type="button"
                  className="btn btn-primary mt-4"
                  disabled={!server.get('tlsPort')}
                  onClick={this.onGoSecureClick()}
                >
                  Go secure
                </button>
              </React.Fragment>
            )}
          </div>
        </div>
      </div>
    );
  };
}

const mapToProps = state => ({
  connected: state.getIn(['connection', 'connected']),
  connecting: state.getIn(['connection', 'connecting']),
  port: state.getIn(['connection', 'port']),
  disconnectionCode: state.getIn(['connection', 'code']),
  server: state.getIn(['server']),
  app: state.getIn(['app']),
});

export default connect(mapToProps)(Landing);
