import { combineReducers } from 'redux-immutable';

import app from './app';
import connection from './connection';
import connections from './connections';

const rootReducer = combineReducers({
  app,
  connection,
  connections,
});

export default rootReducer;
