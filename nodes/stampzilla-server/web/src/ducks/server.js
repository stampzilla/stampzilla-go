import { Map } from 'immutable';
import { defineAction } from 'redux-define';

const c = defineAction(
  'server',
  ['UPDATE'],
);

const defaultState = Map({
});

export const update = state => (
  { type: c.UPDATE, state }
);

export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.UPDATE: {
      return state
        .mergeDeep(action.state);
    }
    default: {
      return state;
    }
  }
}
