import { Map, fromJS } from 'immutable';
import { defineAction } from 'redux-define';

const c = defineAction('persons', ['UPDATE']);

const defaultState = Map({
  list: Map(),
});

// Actions
export function update(connections) {
  return { type: c.UPDATE, connections };
}

// Subscribe to channels and register the action for the packages
export function subscribe(dispatch) {
  return {
    persons: (persons) => dispatch(update(persons)),
  };
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.UPDATE: {
      return state.set('list', fromJS(action.connections));
    }
    default:
      return state;
  }
}
