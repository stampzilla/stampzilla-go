import { combineReducers } from 'redux-immutable';

import app from './app';
import certificates from './certificates';
import cloud from './cloud';
import connection from './connection';
import connections from './connections';
import destinations from './destinations';
import devices from './devices';
import nodes from './nodes';
import persons from './persons';
import requests from './requests';
import rules from './rules';
import savedstates from './savedstates';
import schedules from './schedules';
import senders from './senders';
import server from './server';

const rootReducer = combineReducers({
  app,
  certificates,
  cloud,
  connection,
  connections,
  destinations,
  devices,
  nodes,
  persons,
  requests,
  rules,
  savedstates,
  schedules,
  senders,
  server,
});

export default rootReducer;
