import { createStore, applyMiddleware, compose } from 'redux'
import { routerMiddleware } from 'connected-react-router/immutable'
import thunk from 'redux-thunk'
import createHistory from 'history/createBrowserHistory'
import rootReducer from './ducks'
import { Map } from 'immutable'

export const history = createHistory()

const initialState = Map({})
const enhancers = []
const middleware = [thunk, routerMiddleware(history)]

if (process.env.NODE_ENV === 'development') {
  const devToolsExtension = window.__REDUX_DEVTOOLS_EXTENSION__

  if (typeof devToolsExtension === 'function') {
    enhancers.push(devToolsExtension())
  }
}

const composedEnhancers = compose(
  applyMiddleware(...middleware),
  ...enhancers
)

export default createStore(
  rootReducer(history),
  initialState,
  composedEnhancers
)
