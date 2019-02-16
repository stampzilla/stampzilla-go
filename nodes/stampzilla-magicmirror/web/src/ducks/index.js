import forecast from './forecast'
import devices from './devices'
import config from './config'
import { connectRouter } from 'connected-react-router/immutable'
import { combineReducers } from 'redux-immutable'

export default history =>
  combineReducers({
    forecast,
    devices,
    config,
    router: connectRouter(history)
  })
