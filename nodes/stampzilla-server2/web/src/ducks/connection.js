import { Map, List } from 'immutable';

const CONNECTED   = 'connection/CONNECTED';
const DISCONNECTED = 'connection/DISCONNECTED';
const ERROR = 'connection/ERROR';
const RECEIVED = 'connection/RECEIVED';

const defaultState = Map({
  connected: false,
  error: null,
  messages: List(),
});

// Actions
export function connected() {
  return { type: CONNECTED };
}

export function disconnected() {
  return { type: DISCONNECTED };
}

export function error(error) {
  return { type: ERROR, error };
}

export function received(message) {
  return { type: RECEIVED, message };
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case CONNECTED: {
      return state
        .set('error', null)
        .set('connected', true);
    }
    case DISCONNECTED: {
      return state
        .set('connected', false);
    }
    case ERROR: {
      return state
        .set('error', action.error);
    }
    case RECEIVED: {
      return state
        .set('messages', state.get('messages').push(action.message));
    }
    default: return state;
  }
}
