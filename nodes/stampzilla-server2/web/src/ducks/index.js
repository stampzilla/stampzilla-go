import { combineReducers } from 'redux-immutable';

import app from './app';
import connection from './connection';
import connections from './connections';
import nodes from './nodes';
import server from './server';

const rootReducer = combineReducers({
  app,
  connection,
  connections,
  nodes,
  server,
});

export default rootReducer;
