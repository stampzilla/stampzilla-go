import { Map, fromJS } from 'immutable';
import { defineAction } from 'redux-define';
import makeUUID from 'uuid/v4';

const c = defineAction('persons', ['ADD', 'SAVE', 'REMOVE', 'UPDATE']);

const defaultState = Map({
  list: Map(),
});

// Actions
export function add(person) {
  return { type: c.ADD, person };
}
export function save(person) {
  return { type: c.SAVE, person };
}
export function remove(uuid) {
  return { type: c.REMOVE, uuid };
}
export function update(list) {
  return { type: c.UPDATE, list };
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
    case c.ADD: {
      const person = {
        ...action.person,
        uuid: makeUUID(),
      };
      return state.setIn(['list', person.uuid], fromJS(person));
    }
    case c.SAVE: {
      return state.mergeIn(['list', action.person.uuid], fromJS(action.person));
    }
    case c.REMOVE: {
      return state.deleteIn(['list', action.uuid]);
    }
    case c.UPDATE: {
      return state.set('list', fromJS(action.list));
    }
    default:
      return state;
  }
}
