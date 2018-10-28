import { combineReducers } from 'redux-immutable';

import connection from './connection';
import connections from './connections';

const rootReducer = combineReducers({
  connection,
  connections,
})

export default rootReducer;
