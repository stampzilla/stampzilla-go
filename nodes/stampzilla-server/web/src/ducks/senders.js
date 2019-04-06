import { Map, fromJS } from 'immutable';
import { defineAction } from 'redux-define';
import makeUUID from 'uuid/v4';

const c = defineAction(
  'senders',
  ['ADD', 'SAVE', 'UPDATE', 'UPDATE_STATE'],
);

const defaultState = Map({
  list: Map(),
});

// Actions
export function add(sender) {
  return { type: c.ADD, sender };
}
export function save(sender) {
  return { type: c.SAVE, sender };
}
export function update(senders) {
  return { type: c.UPDATE, senders };
}
export function updateState(senders) {
  return { type: c.UPDATE_STATE, senders };
}

// Subscribe to channels and register the action for the packages
export function subscribe(dispatch) {
  return {
    senders: senders => dispatch(update(senders)),
  };
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.ADD: {
      const sender = {
        ...action.sender,
        uuid: makeUUID(),
      };
      return state
        .setIn(['list', sender.uuid], fromJS(sender));
    }
    case c.SAVE: {
      return state
        .mergeIn(['list', action.sender.uuid], fromJS(action.sender));
    }
    case c.UPDATE: {
      return state
        .set('list', fromJS(action.senders));
    }
    case c.UPDATE_STATE: {
      return state
        .set('state', fromJS(action.senders));
    }
    default: return state;
  }
}
