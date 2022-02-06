import { Map, fromJS } from 'immutable';
import { defineAction } from 'redux-define';
import { v4 as makeUUID } from 'uuid';

const c = defineAction(
  'schedules',
  ['ADD', 'SAVE', 'REMOVE', 'UPDATE', 'UPDATE_STATE'],
);

const defaultState = Map({
  list: Map(),
  state: Map(),
});

// Actions
export function add(schedule) {
  return { type: c.ADD, schedule };
}
export function save(schedule) {
  return { type: c.SAVE, schedule };
}
export function remove(uuid) {
  return { type: c.REMOVE, uuid };
}
export function update(schedules) {
  return { type: c.UPDATE, schedules };
}
export function updateState(schedules) {
  return { type: c.UPDATE_STATE, schedules };
}

// Subscribe to channels and register the action for the packages
export function subscribe(dispatch) {
  return {
    schedules: (schedules) => dispatch(update(schedules)),
    server: ({ schedules }) => schedules && dispatch(updateState(schedules)),
  };
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.ADD: {
      const schedule = {
        ...action.schedule,
        uuid: makeUUID(),
      };
      return state
        .setIn(['list', schedule.uuid], fromJS(schedule));
    }
    case c.SAVE: {
      return state
        .mergeIn(['list', action.schedule.uuid], fromJS(action.schedule));
    }
    case c.REMOVE: {
      return state.deleteIn(['list', action.uuid]);
    }
    case c.UPDATE: {
      return state
        .set('list', fromJS(action.schedules));
    }
    case c.UPDATE_STATE: {
      return state
        .set('state', fromJS(action.schedules));
    }
    default: return state;
  }
}
