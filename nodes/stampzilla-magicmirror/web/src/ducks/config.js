import { Map, fromJS } from 'immutable'
import { defineAction } from 'redux-define'

const c = defineAction('config', ['UPDATE'])

const defaultState = Map({})

// Actions
export function update(config) {
  return { type: c.UPDATE, config }
}

// Subscribe to channels and register the action for the packages
export function subscribe(dispatch) {
  return {
    config: config => dispatch(update(config))
  }
}

// Reducer
export default function reducer(state = defaultState, action) {
  switch (action.type) {
    case c.UPDATE: {
      return fromJS(action.config)
    }
    default:
      return state
  }
}
