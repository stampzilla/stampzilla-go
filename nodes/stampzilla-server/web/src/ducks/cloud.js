import { Map, fromJS } from 'immutable';
import { defineAction } from 'redux-define';

const c = defineAction('cloud', ['UPDATE']);

const defaultState = Map({
  config: Map(),
  state: Map(),
});

// Actions
export function update(data) {
  return { type: c.UPDATE, data };
}

// Subscribe to channels and register the action for the packages
export function subscribe(dispatch) {
  return {
    cloud: (data) => dispatch(update(data)),
  };
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.UPDATE: {
      return state
        .set('config', fromJS(action.data.config))
        .set('state', fromJS(action.data.state));
    }
    default:
      return state;
  }
}
