import { Map } from 'immutable';
import { defineAction } from 'redux-define';

const c = defineAction(
  'app',
  ['UPDATE'],
);

const defaultState = Map({
  url: `${window.location.protocol.match(/^https/) ? 'wss' : 'ws'}://${window.location.host}/ws`,
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
