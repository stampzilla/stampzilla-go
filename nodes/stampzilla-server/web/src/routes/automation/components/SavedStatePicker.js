import { connect } from 'react-redux';
import { v4 as makeUUID } from 'uuid';
import Modal from 'react-modal';
import React from 'react';
import classnames from 'classnames';

import './SavedStatePicker.scss';
import { save } from '../../../ducks/savedstates';
import Scene from './Scene';

Modal.setAppElement('#app');

const valueToText = (value, savedstates) => {
  if (value && value.length === 36) {
    return (
      <React.Fragment>
        <small>Scene</small>
        {' '}
        <strong>{savedstates.getIn([value, 'name']) || value}</strong>
      </React.Fragment>
    );
  }

  if (value && value.length > 0) {
    return (
      <React.Fragment>
        <small>Delay for</small>
        {' '}
        <strong>{value}</strong>
      </React.Fragment>
    );
  }

  return 'New action';
};

const valueToTab = (value) => {
  if (value && value.length === 36) {
    return 'state';
  }

  if (value && value.length > 0) {
    return 'time';
  }

  return 'state';
};

class SavedStatePicker extends React.Component {
  constructor(props) {
    super(props);

    const scene = props.savedstates.get(props.value);
    this.state = {
      tab: valueToTab(props.value),
      value: props.value,
      modalIsOpen: props.value === undefined,
      scene: (scene && scene.toJS()) || {},
    };

    this.selectRef = React.createRef();
  }

  componentWillReceiveProps(props) {
    if (props.value !== this.props.value) {
      const scene = props.savedstates.get(props.value);
      this.setState({
        tab: valueToTab(props.value),
        value: props.value,
        scene: (scene && scene.toJS()) || {},
      });
    }
  }

  onSceneChange = () => (state) => {
    const { scene } = this.state;
    this.setState({
      scene: { ...scene, state },
    });
  };

  openModal = () => () => {
    const { value } = this.props;
    this.setState({
      modalIsOpen: true,
      tab: valueToTab(value),
      value,
    });
    return false;
  };

  closeModal = () => () => {
    this.setState({ modalIsOpen: false });
  };

  save = () => () => {
    const { onChange, dispatch } = this.props;
    const { value, scene } = this.state;

    if (value.length === 36) {
      dispatch(save({ ...scene, uuid: value }));
    }

    onChange(value);
    this.setState({ modalIsOpen: false });
  };

  render() {
    const {
      tab, modalIsOpen, value, scene,
    } = this.state;
    const { state: states } = scene || {};
    const { savedstates, options } = this.props;

    return (
      <React.Fragment>
        <div
          className="btn btn-secondary btn-block"
          onClick={this.openModal()}
          role="button"
          tabIndex={-1}
        >
          {valueToText(this.props.value, savedstates)}
        </div>
        <Modal
          className="Modal__Bootstrap modal-dialog saved-state-modal"
          closeTimeoutMS={150}
          isOpen={modalIsOpen}
          onRequestClose={this.closeModal()}
        >
          <div className="modal-content">
            <div style={{ display: 'flex' }}>
              <ul
                className="nav nav-tabs"
                style={{ marginLeft: '-1px', flex: '1 1 0%' }}
              >
                <li className="nav-item">
                  <a
                    className={classnames(
                      'nav-link',
                      tab === 'state' && 'active',
                    )}
                    role="button"
                    tabIndex="0"
                    onClick={() => this.setState({ tab: 'state' })}
                  >
                    Scene
                  </a>
                </li>
                {!(value && value.length === 36) && !options.hideDelay && (
                  <li className="nav-item">
                    <a
                      className={classnames(
                        'nav-link',
                        tab === 'time' && 'active',
                      )}
                      role="button"
                      tabIndex="0"
                      onClick={() => this.setState({ tab: 'time' })}
                    >
                      Time delay
                    </a>
                  </li>
                )}
              </ul>
              <button
                type="button"
                className="close"
                style={{
                  padding: '0px 15px',
                  borderBottom: '1px solid rgb(222, 226, 230)',
                }}
                onClick={this.closeModal()}
              >
                <span aria-hidden="true">×</span>
                <span className="sr-only">Close</span>
              </button>
            </div>
            <div className="modal-body">
              {tab === 'time' && (
                <form>
                  <div className="form-group">
                    <label htmlFor="name">Duration</label>
                    <input
                      type="text"
                      id="name"
                      className="form-control"
                      placeholder="ex. 2h15m"
                      value={value || ''}
                      onChange={event => this.setState({ value: event.target.value })
                      }
                    />
                    <small className="form-text text-muted">
                      ParseDuration parses a duration string. A duration string
                      is a sequence of decimal numbers, each with optional
                      fraction and a unit suffix, such as
                      {' '}
                      <strong>300ms</strong>
                      ,
                      {' '}
                      <strong>1.5h</strong>
                      {' '}
or
                      {' '}
                      <strong>2h45m</strong>
. Valid
                      time units are
                      {' '}
                      <strong>n</strong>
,
                      {' '}
                      <strong>us</strong>
                      {' '}
(or
                      {' '}
                      <strong>µs</strong>
                      ),
                      <strong>ms</strong>
,
                      <strong>s</strong>
,
                      <strong>m</strong>
,
                      <strong>h</strong>
.
                    </small>
                  </div>
                </form>
              )}
              {tab === 'state' && (!value || value.length !== 36) && (
                <React.Fragment>
                  <button
                    type="button"
                    className="btn btn-success btn-block"
                    onClick={() => this.setState({ value: makeUUID() })}
                  >
                    Create a new scene
                  </button>
                  <div className="text-center p-4">or select an existing</div>
                  <div>
                    <select
                      size={10}
                      style={{ width: '100%' }}
                      ref={this.selectRef}
                    >
                      {savedstates.map(savedstate => (
                        <option value={savedstate.get('uuid')}>
                          {savedstate.get('name')}
                        </option>
                      ))}
                    </select>
                  </div>
                  <button
                    type="button"
                    className="btn btn-secondary btn-block"
                    onClick={() => {
                      this.setState({
                        value: this.selectRef.current.value,
                        scene:
                          savedstates
                            .get(this.selectRef.current.value)
                            .toJS() || {},
                      });
                    }}
                  >
                    Select
                  </button>
                </React.Fragment>
              )}
              {tab === 'state' && value && value.length === 36 && (
                <form>
                  <div className="form-group">
                    <label htmlFor="name">Name</label>
                    <input
                      type="text"
                      id="name"
                      className="form-control"
                      placeholder=""
                      value={(scene && scene.name) || ''}
                      onChange={event => this.setState({
                        scene: { ...scene, name: event.target.value },
                      })
                      }
                    />
                    <small className="form-text text-muted">
                      Give the state a good name so you can find it later
                    </small>
                  </div>
                  <div className="form-group">
                    <label htmlFor="name">State</label>
                    <Scene
                      id={value}
                      onChange={this.onSceneChange()}
                      states={states}
                    />
                  </div>
                </form>
              )}
            </div>
            <div className="modal-footer">
              {!value && (
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={this.closeModal()}
                >
                  Close
                </button>
              )}
              {value && (
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={this.save()}
                >
                  Use
                </button>
              )}
            </div>
          </div>
        </Modal>
      </React.Fragment>
    );
  }
}

const mapStateToProps = state => ({
  devices: state.getIn(['devices', 'list']),
  savedstates: state.getIn(['savedstates', 'list']),
});

export const ConnectedSavedStatePicker = connect(mapStateToProps)(
  SavedStatePicker,
);
const SavedStateWidget = props => <ConnectedSavedStatePicker {...props} />;
export default SavedStateWidget;
