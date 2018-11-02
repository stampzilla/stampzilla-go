import { combineReducers } from 'redux-immutable';

import app from './app';
import connection from './connection';
import connections from './connections';
import server from './server';

const rootReducer = combineReducers({
  app,
  connection,
  connections,
  server,
});

export default rootReducer;
