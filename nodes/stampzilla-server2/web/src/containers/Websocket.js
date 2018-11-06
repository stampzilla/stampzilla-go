import { Component } from 'react';
import { connect } from 'react-redux';
import ReconnectableWebSocket from 'reconnectable-websocket';
import Url from 'url';

import {
  connected,
  disconnected,
  received,
} from '../ducks/connection';
import { update as updateConnections } from '../ducks/connections';
import { update as updateServer } from '../ducks/server';

// Placeholder until we have the write func from the websocket
let writeFunc = () => {
  throw new Error('Not initialized yet');
};

class Websocket extends Component {
  componentDidMount() {
    this.setupSocket(this.props);
  }

  componentWillReceiveProps(props) {
    this.setupSocket(props);
  }

  componentWillUnmount() {
    if (typeof this.socket !== 'undefined') {
      this.socket.close();
    }
  }

  onOpen = () => () => {
    this.props.dispatch(connected());

    const url = Url.parse(this.props.url);
    if (url.protocol === 'wss:') {
      this.props.dispatch(updateServer({ secure: true }));
    }
  }
  onClose = () => () => {
    this.props.dispatch(disconnected());
  }
  onMessage = () => (event) => {
    const { dispatch } = this.props;
    const parsed = JSON.parse(event.data);

    dispatch(received(parsed));
    switch (parsed.type) {
      case 'server-info': {
        dispatch(updateServer(parsed.body));
        break;
      }
      case 'connections': {
        dispatch(updateConnections(parsed.body));
        break;
      }
      default: {
        break;
      }
    }
  }

  setupSocket(props) {
    const { url } = props;
    if (this.socket) {
      // this is becuase there is a bug in reconnecting websocket causing it to retry forever
      this.socket.onclose = () => throw ('force close socket');
      // Close the existing connection
      this.socket.close();
    }

    this.socket = new ReconnectableWebSocket(url, ['gui'], {
      reconnectInterval: 3000,
      timeoutInterval: 1000,
    });
    writeFunc = this.socket.send;
    this.socket.onmessage = this.onMessage();
    this.socket.onopen = this.onOpen();
    this.socket.onclose = this.onClose();
  }

  render = () => null;
}

const mapToProps = state => ({
  url: state.getIn(['app', 'url']),
});
export default connect(mapToProps)(Websocket);

export const write = msg => writeFunc(JSON.stringify(msg));
