import { Map } from 'immutable';
import { defineAction } from 'redux-define';

const c = defineAction(
  'connection',
  ['CONNECTED', 'DISCONNECTED', 'ERROR'],
);

const defaultState = Map({
  connected: null,
  error: null,
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

// Reducer
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
    default: return state;
  }
}
