import { Map, fromJS } from 'immutable';
import { defineAction } from 'redux-define';
import makeUUID from 'uuid/v4';

const c = defineAction(
  'schedules',
  ['ADD', 'SAVE', 'UPDATE'],
);

const defaultState = Map({
  list: Map(),
});

// Actions
export function add(schedule) {
  return { type: c.ADD, schedule };
}
export function save(schedule) {
  return { type: c.SAVE, schedule };
}
export function update(schedules) {
  return { type: c.UPDATE, schedules };
}

// Subscribe to channels and register the action for the packages
export function subscribe(dispatch) {
  return {
    schedules: schedules => dispatch(update(schedules)),
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
    case c.UPDATE: {
      return state
        .set('list', fromJS(action.schedules));
    }
    default: return state;
  }
}
