import forecast from './forecast'
import { connectRouter } from 'connected-react-router/immutable'
import { combineReducers } from 'redux-immutable'

export default history =>
  combineReducers({
    forecast,
    router: connectRouter(history)
  })
