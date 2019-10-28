import { write } from '../components/Websocket';

const senders = (store) => (next) => (action) => {
  const prev = store.getState().getIn(['senders', 'list']);
  const result = next(action);
  const after = store.getState().getIn(['senders', 'list']);

  if (!after.equals(prev) && action.type !== 'senders_UPDATE') {
    write({
      type: 'update-senders',
      body: after.toJS(),
    });
  }
  return result;
};

export default senders;
