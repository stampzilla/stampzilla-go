import { applyMiddleware, createStore, compose } from 'redux';
import ReduxThunk from 'redux-thunk';
import rootReducer from './ducks';

const middleware = [
  ReduxThunk,
];

const preloadedState = undefined;

const composeEnhancers =
   (typeof window !== 'undefined' && window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__) || compose

const newStore = (initialState, extraMiddlewares = []) => createStore(
  rootReducer,
  initialState,
  composeEnhancers(applyMiddleware(
    ...extraMiddlewares,
    ...middleware,
  )),
);
const store = newStore(preloadedState);

export default store;
export { newStore };
