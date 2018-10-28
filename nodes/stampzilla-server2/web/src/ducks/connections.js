import { Map, List } from 'immutable';

const UPDATE = 'connections/UPDATE';

const defaultState = Map({
  list: Map(),
});

// Actions
export function update(connections) {
  return { type: UPDATE, connections };
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case UPDATE: {
      return state
        .set('list', Map(action.connections))
    }
    default: return state;
  }
}
