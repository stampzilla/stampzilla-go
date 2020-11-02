import { List, Map } from 'immutable';
import { defineAction } from 'redux-define';

const c = defineAction('connection', [
  'CONNECTING',
  'CONNECTED',
  'DISCONNECTED',
  'ERROR',
]);

const defaultState = Map({
  trying: false,
  connected: null,
  method: null,
  port: 0,
  code: 0,
  reason: '',
  error: null,
  messages: List(),
  user: null,
});

// Actions
export function connecting(port) {
  return { type: c.CONNECTING, port };
}
export function connected(port, method, user) {
  return {
    type: c.CONNECTED,
    port,
    method,
    user,
  };
}

export function disconnected(code, reason, retrying) {
  return (dispatch, getState) => {
    if (
      getState().getIn(['connection', 'connected']) !== false
      || (code && getState().getIn(['connection', 'code']) !== code)
    ) {
      dispatch({
        type: c.DISCONNECTED,
        code,
        reason,
        retrying,
      });
    }
  };
}

export function error(err) {
  return { type: c.ERROR, error: err };
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.CONNECTING: {
      return state.set('connecting', true).set('port', action.port);
    }
    case c.CONNECTED: {
      return state
        .set('error', null)
        .set('connected', true)
        .set('method', action.method)
        .set('user', action.user)
        .set('code', 0)
        .set('port', action.port)
        .set('reason', '');
    }
    case c.DISCONNECTED: {
      return state
        .set('connected', false)
        .set('method', null)
        .set('user', null)
        .set('code', action.code)
        .set('connecting', action.retrying)
        .set('reason', action.reason);
    }
    case c.ERROR: {
      return state.set('error', action.error);
    }
    default:
      return state;
  }
}
