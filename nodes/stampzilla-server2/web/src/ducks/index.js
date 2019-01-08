import { combineReducers } from 'redux-immutable';

import app from './app';
import certificates from './certificates';
import connection from './connection';
import connections from './connections';
import devices from './devices';
import nodes from './nodes';
import requests from './requests';
import server from './server';

const rootReducer = combineReducers({
  app,
  certificates,
  connection,
  connections,
  devices,
  nodes,
  requests,
  server,
});

export default rootReducer;
