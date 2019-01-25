import { List, Map } from 'immutable';
import { defineAction } from 'redux-define';

const c = defineAction(
  'connection',
  ['CONNECTED', 'DISCONNECTED', 'ERROR', 'RECEIVED'],
);

const defaultState = Map({
  connected: null,
  error: null,
  messages: List(),
});

// Actions
export function connected() {
  return { type: c.CONNECTED };
}

export function disconnected() {
  return (dispatch, getState) => {
    if (getState().getIn(['connection', 'connected']) !== false) {
      dispatch({ type: c.DISCONNECTED });
    }
  };
}

export function error(err) {
  return { type: c.ERROR, error: err };
}

export function received(message) {
  return { type: c.RECEIVED, message };
}

// Reducer
let idCount = 0;
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.CONNECTED: {
      return state
        .set('error', null)
        .set('connected', true);
    }
    case c.DISCONNECTED: {
      return state
        .set('connected', false);
    }
    case c.ERROR: {
      return state
        .set('error', action.error);
    }
    case c.RECEIVED: {
      idCount += 1;
      return state
        .set('messages', state.get('messages').push({
          ...action.message,
          id: idCount,
        }).slice(-10));
    }
    default: return state;
  }
}
