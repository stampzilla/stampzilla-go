import { Map, fromJS } from 'immutable';
import { defineAction } from 'redux-define';
import { v4 as makeUUID } from 'uuid';

const c = defineAction('savedstates', ['ADD', 'SAVE', 'UPDATE']);

const defaultState = Map({
  list: Map(),
});

// Actions
export function add(states) {
  return { type: c.ADD, states };
}
export function save(state) {
  return { type: c.SAVE, state };
}
export function update(states) {
  return { type: c.UPDATE, states };
}

// Subscribe to channels and register the action for the packages
export function subscribe(dispatch) {
  return {
    savedstates: states => dispatch(update(states)),
  };
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.ADD: {
      const s = {
        ...action.state,
        uuid: makeUUID(),
      };
      return state.setIn(['list', s.uuid], fromJS(s));
    }
    case c.SAVE: {
      return state.mergeIn(['list', action.state.uuid], fromJS(action.state));
    }
    case c.UPDATE: {
      return state.set('list', fromJS(action.states));
    }
    default:
      return state;
  }
}
