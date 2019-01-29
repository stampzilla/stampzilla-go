import { write } from '../components/Websocket';

const savedstates = store => next => (action) => {
  const prev = store.getState().getIn(['savedstates', 'list']);
  const result = next(action);
  const after = store.getState().getIn(['savedstates', 'list']);

  if (!after.equals(prev) && action.type !== 'savedstates_UPDATE') {
    write({
      type: 'update-savedstates',
      body: after.toJS(),
    });
  }
  return result;
};

export default savedstates;
