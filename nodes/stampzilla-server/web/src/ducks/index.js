import { combineReducers } from 'redux-immutable';

import app from './app';
import certificates from './certificates';
import connection from './connection';
import connections from './connections';
import devices from './devices';
import nodes from './nodes';
import requests from './requests';
import rules from './rules';
import savedstates from './savedstates';
import server from './server';
import schedules from './schedules';

const rootReducer = combineReducers({
  app,
  certificates,
  connection,
  connections,
  devices,
  nodes,
  requests,
  rules,
  savedstates,
  server,
  schedules,
});

export default rootReducer;
