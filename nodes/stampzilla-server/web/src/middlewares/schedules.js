import { write } from '../components/Websocket';

const schedules = store => next => (action) => {
  const prev = store.getState().getIn(['schedules', 'list']);
  const result = next(action);
  const after = store.getState().getIn(['schedules', 'list']);

  if (!after.equals(prev) && action.type !== 'schedules_UPDATE') {
    write({
      type: 'update-schedules',
      body: after.toJS(),
    });
  }
  return result;
};

export default schedules;
