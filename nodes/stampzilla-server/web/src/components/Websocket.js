import { Component } from 'react';
import { connect } from 'react-redux';
import ReconnectingWebSocket from 'reconnecting-websocket';
import Url from 'url';

import { subscribe as certificates } from '../ducks/certificates';
import { connected, disconnected, connecting } from '../ducks/connection';
import { subscribe as cloud } from '../ducks/cloud';
import { subscribe as connections } from '../ducks/connections';
import { subscribe as destinations } from '../ducks/destinations';
import { subscribe as devices } from '../ducks/devices';
import { subscribe as nodes } from '../ducks/nodes';
import { subscribe as persons } from '../ducks/persons';
import { subscribe as requests } from '../ducks/requests';
import { subscribe as rules } from '../ducks/rules';
import { subscribe as savedstates } from '../ducks/savedstates';
import { subscribe as schedules } from '../ducks/schedules';
import { subscribe as senders } from '../ducks/senders';
import { update as updateServer } from '../ducks/server';

// Placeholder until we have the write func from the websocket
let writeSocket = null;
const writeFunc = (data) => {
  if (writeSocket === null) {
    throw new Error('Not initialized yet');
  }
  writeSocket.send(data);
};

let requestId = 0;
let activeRequests = [];
export const write = (msg) => writeFunc(JSON.stringify(msg));
export const request = (msg) => new Promise((resolve, reject) => {
  activeRequests.push({
    id: requestId,
    resolve,
    reject,
  });

  writeFunc(
    JSON.stringify({
      ...msg,
      request: requestId,
    }),
  );
  requestId += 1;
});

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

  onOpen = ({ method, user }) => {
    const url = Url.parse(this.props.url);
    this.props.dispatch(connected(url.port, method, user));

    if (url.protocol === 'wss:') {
      this.props.dispatch(updateServer({ secure: true }));
    }

    this.subscribe({
      cloud,
      certificates,
      connections,
      destinations,
      devices,
      nodes,
      persons,
      requests,
      rules,
      savedstates,
      schedules,
      senders,
    });
  };

  onClose = (err) => {
    let retrying = true;
    if (err.code === 4001) {
      this.socket.close();
      retrying = false;
    }

    this.props.dispatch(disconnected(err.code, err.reason, retrying));
    this.subscriptions = {};
  };

  onMessage = (event) => {
    const { dispatch } = this.props;
    const parsed = JSON.parse(event.data);

    const subscriptions = this.subscriptions[parsed.type];
    if (subscriptions) {
      subscriptions.forEach((callback) => callback(parsed.body));
    }
    switch (parsed.type) {
      case 'server-info': {
        dispatch(updateServer(parsed.body));
        break;
      }
      case 'ready': {
        this.onOpen(parsed.body);
        break;
      }
      case 'success': {
        const req = activeRequests.find((a) => a.id === parsed.request);
        req.resolve(parsed.body);
        activeRequests = activeRequests.filter((a) => a.id !== parsed.request);
        break;
      }
      case 'failure': {
        const req = activeRequests.find((a) => a.id === parsed.request);
        req.reject(parsed.body);
        activeRequests = activeRequests.filter((a) => a.id !== parsed.request);
        break;
      }
      default: {
        // Nothing
      }
    }
  };

  setupSocket(props) {
    const { url } = props;
    if (this.socket) {
      // this is becuase there is a bug in reconnecting websocket causing it to retry forever
      this.socket.onclose = () => {
        throw new Error('force close socket');
      };
      // Close the existing connection
      this.socket.close();
    }

    const secureUrl = Url.format({
      protocol: 'wss:',
      hostname: window.location.hostname,
      port: window.location.port,
      pathname: '/ws',
    });

    this.socket = new ReconnectingWebSocket(
      window.location.protocol === 'https' ? secureUrl : url,
      ['gui'],
      {
        reconnectInterval: 3000,
        timeoutInterval: 1000,
      },
    );

    const parsedUrl = Url.parse(url);
    props.dispatch(connecting(parsedUrl.port));

    writeSocket = this.socket;
    // writeFunc = this.socket.send;
    this.socket.onmessage = this.onMessage;
    // this.socket.onopen = this.onOpen;
    this.socket.onerror = this.onClose;
    this.socket.onclose = this.onClose;
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
  };

  render = () => null;
}

const mapToProps = (state) => ({
  url: state.getIn(['app', 'url']),
});
export default connect(mapToProps)(Websocket);
