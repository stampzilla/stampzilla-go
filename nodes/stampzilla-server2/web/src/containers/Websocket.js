import { Component } from 'react';
import { connect } from 'react-redux';
import ReconnectableWebSocket from 'reconnectable-websocket';

import {
  connected,
  disconnected,
  error,
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
    const { url } = this.props;
    this.socket = new ReconnectableWebSocket(url, ['gui'], {
      reconnectInterval: 3000,
      timeoutInterval: 1000,
    });
    writeFunc = this.socket.send;
    this.socket.onmessage = this.onMessage();
    this.socket.onopen = this.onOpen();
    this.socket.onclose = this.onClose();
    this.socket.onerror = this.onError();
  }

  componentWillUnmount() {
    this.socket.close();
  }

  onOpen = () => () => {
    this.props.dispatch(connected());
  }
  onClose = () => () => {
    this.props.dispatch(disconnected());
  }
  onError = () => (err) => {
    this.props.dispatch(error(err));
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


  render = () => null;
}

const mapToProps = state => ({
  url: state.getIn(['app', 'url']),
});
export default connect(mapToProps)(Websocket);

export const write = msg => writeFunc(JSON.stringify(msg));
