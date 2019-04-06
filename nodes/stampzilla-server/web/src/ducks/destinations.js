import { Map, fromJS } from 'immutable';
import { defineAction } from 'redux-define';
import makeUUID from 'uuid/v4';

const c = defineAction('destinations', [
  'ADD',
  'SAVE',
  'UPDATE',
  'UPDATE_STATE',
]);

const defaultState = Map({
  list: Map(),
});

// Actions
export function add(destination) {
  return { type: c.ADD, destination };
}
export function save(destination) {
  return { type: c.SAVE, destination };
}
export function update(destinations) {
  return { type: c.UPDATE, destinations };
}
export function updateState(destinations) {
  return { type: c.UPDATE_STATE, destinations };
}

// Subscribe to channels and register the action for the packages
export function subscribe(dispatch) {
  return {
    destinations: destinations => dispatch(update(destinations)),
  };
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.ADD: {
      const destination = {
        ...action.destination,
        uuid: makeUUID(),
      };
      return state.setIn(['list', destination.uuid], fromJS(destination));
    }
    case c.SAVE: {
      return state.mergeIn(
        ['list', action.destination.uuid],
        fromJS(action.destination),
      );
    }
    case c.UPDATE: {
      return state.set('list', fromJS(action.destinations));
    }
    case c.UPDATE_STATE: {
      return state.set('state', fromJS(action.destinations));
    }
    default:
      return state;
  }
}
