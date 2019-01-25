import { applyMiddleware, compose, createStore } from 'redux';
import Immutable from 'immutable';
import ReduxThunk from 'redux-thunk';
import persistState from 'redux-localstorage';

import rootReducer from './ducks';
import rules from './middlewares/rules';
import schedules from './middlewares/schedules';

const middleware = [
  ReduxThunk,
  rules,
  schedules,
];

const preloadedState = undefined;

const composeEnhancers =
   (typeof window !== 'undefined' && window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__) || compose; // eslint-disable-line no-underscore-dangle

const localStorageConfig = {
  slicer: paths => state => (paths ? state.filter((v, k) => paths.indexOf(k) > -1) : state),
  serialize: subset => JSON.stringify(subset.toJS()),
  deserialize: serializedData => Immutable.fromJS(JSON.parse(serializedData)) || Immutable.Map(),
  merge: (initialState, persistedState) =>
    (initialState ? initialState.mergeDeep(persistedState) : persistedState),
};

const newStore = (initialState, extraMiddlewares = []) => createStore(
  rootReducer,
  initialState,
  composeEnhancers(
    applyMiddleware(
      ...extraMiddlewares,
      ...middleware,
    ),
    persistState(['app'], localStorageConfig),
  ),
);
const store = newStore(preloadedState);

export default store;
export { newStore };
