import { Component } from 'react';
import { connect } from 'react-redux';
import ReconnectableWebSocket from 'reconnectingwebsocket';
import Url from 'url';

import { subscribe as certificates } from '../ducks/certificates';
import {
  connected,
  disconnected,
  received,
} from '../ducks/connection';
import { subscribe as connections } from '../ducks/connections';
import { subscribe as devices } from '../ducks/devices';
import { subscribe as nodes } from '../ducks/nodes';
import { subscribe as requests } from '../ducks/requests';
import { subscribe as rules } from '../ducks/rules';
import { subscribe as savedstates } from '../ducks/savedstates';
import { subscribe as schedules } from '../ducks/schedules';
import { update as updateServer } from '../ducks/server';

// Placeholder until we have the write func from the websocket
let writeSocket = null;
const writeFunc = (data) => {
  if (writeSocket === null) {
    throw new Error('Not initialized yet');
  }
  writeSocket.send(data);
};
export const write = msg => writeFunc(JSON.stringify(msg));

class Websocket extends Component {
  constructor(props) {
    super(props);

    this.subscriptions = {};
  }

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

    this.subscribe({
      certificates,
      connections,
      devices,
      nodes,
      requests,
      rules,
      savedstates,
      schedules,
    });
  }
  onClose = () => () => {
    this.props.dispatch(disconnected());
    this.subscriptions = {};
  }
  onMessage = () => (event) => {
    const { dispatch } = this.props;
    const parsed = JSON.parse(event.data);

    dispatch(received(parsed));
    const subscriptions = this.subscriptions[parsed.type];
    if (subscriptions) {
      subscriptions.forEach(callback => callback(parsed.body));
    }
    switch (parsed.type) {
      case 'server-info': {
        dispatch(updateServer(parsed.body));
        break;
      }
      default: {
        // Nothing
      }
    }
  }

  setupSocket(props) {
    const { url } = props;
    if (this.socket) {
      // this is becuase there is a bug in reconnecting websocket causing it to retry forever
      this.socket.onclose = () => { throw new Error('force close socket'); };
      // Close the existing connection
      this.socket.close();
    }

    this.socket = new ReconnectableWebSocket(url, ['gui'], {
      reconnectInterval: 3000,
      timeoutInterval: 1000,
    });
    writeSocket = this.socket;
    //writeFunc = this.socket.send;
    this.socket.onmessage = this.onMessage();
    this.socket.onopen = this.onOpen();
    this.socket.onclose = this.onClose();
  }

  subscribe = (ducks) => {
    const subscriptions = [];
    Object.keys(ducks).forEach((duck) => {
      if (!ducks[duck]) {
        return;
      }

      const channels = ducks[duck](this.props.dispatch);
      Object.keys(channels).forEach((channel) => {
        this.subscriptions[channel] = this.subscriptions[channel] || [];
        this.subscriptions[channel].push(channels[channel]);
        subscriptions.push(channel);
      });
    });

    write({
      type: 'subscribe',
      body: subscriptions,
    });
  }

  render = () => null;
}

const mapToProps = state => ({
  url: state.getIn(['app', 'url']),
});
export default connect(mapToProps)(Websocket);

