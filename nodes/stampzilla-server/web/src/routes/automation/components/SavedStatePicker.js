import React from 'react';
import Modal from 'react-modal';
import classnames from 'classnames';
import { connect } from 'react-redux';

import './SavedStatePicker.scss';
import Scene from './Scene';

Modal.setAppElement('#app');

const valueToText = (value) => {
  if (value && value.length === 36) {
    return `Scene ${value}`;
  }

  if (value && value.length > 0) {
    return `Delay for ${value}`;
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
    this.state = {
      tab: valueToTab(props.value),
      value: props.value,
      modalIsOpen: props.value === undefined,
    };
  }

  componentWillReceiveProps(props) {
    if (props.value !== this.props.value) {
      this.setState({
        tab: valueToTab(props.value),
        value: props.value,
      });
    }
  }

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
    const { onChange } = this.props;
    const { value } = this.state;
    onChange(value);
    this.setState({ modalIsOpen: false });
  };

  render() {
    const { tab, modalIsOpen, value } = this.state;

    return (
      <React.Fragment>
        <div
          className="btn btn-secondary btn-block"
          onClick={this.openModal()}
          role="button"
          tabIndex={-1}
        >
          {valueToText(this.props.value)}
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
                      value={value}
                      onChange={event =>
                        this.setState({ value: event.target.value })
                      }
                    />
                    <small className="form-text text-muted">
                      ParseDuration parses a duration string. A duration string
                      is a sequence of decimal numbers, each with optional
                      fraction and a unit suffix, such as <strong>300ms</strong>
                      , <strong>1.5h</strong> or <strong>2h45m</strong>. Valid
                      time units are <strong>n</strong>, <strong>us</strong> (or{' '}
                      <strong>µs</strong>), <strong>ms</strong>,
                      <strong>s</strong>, <strong>m</strong>, <strong>h</strong>
                      .
                    </small>
                  </div>
                </form>
              )}
              {tab === 'state' && (
                <form>
                  <div className="form-group">
                    <label htmlFor="name">Name</label>
                    <input
                      type="text"
                      id="name"
                      className="form-control"
                      placeholder=""
                      value=""
                    />
                    <small className="form-text text-muted">
                      Give the state a good name so you can find it later
                    </small>
                  </div>
                  <div className="form-group">
                    <label htmlFor="name">State</label>
                    <Scene id={value} />
                  </div>
                </form>
              )}
            </div>
            <div className="modal-footer">
              <button
                type="button"
                className="btn btn-primary"
                onClick={this.save()}
              >
                Use
              </button>
            </div>
          </div>
        </Modal>
      </React.Fragment>
    );
  }
}

const mapStateToProps = state => ({
  devices: state.getIn(['devices', 'list']),
});

export default connect(mapStateToProps)(SavedStatePicker);

/*
<div class="ReactModal__Overlay ReactModal__Overlay--after-open" style="position: fixed; top: 0px; left: 0px; right: 0px; bottom: 0px; background-color: rgba(255, 255, 255, 0.75); z-index: 20000;"><div class="ReactModal__Content ReactModal__Content--after-open Modal__Bootstrap modal-dialog" tabindex="-1" role="dialog" aria-label="Example Modal" style="z-index: 20000;"><div class="modal-content" style="border-top: none rgb(222, 226, 230); border-right-color: rgb(222, 226, 230); border-bottom-color: rgb(222, 226, 230); border-left-color: rgb(222, 226, 230);"><div style="display: flex;"><ul class="nav nav-tabs" style="margin-left: -1px; flex: 1 1 0%;"><li class="nav-item"><a class="nav-link active" role="button" tabindex="0">Scene</a></li><li class="nav-item"><a class="nav-link" role="button" tabindex="0">Time delay</a></li></ul><button type="button" class="close" style="padding: 0px 15px; border-bottom: 1px solid rgb(222, 226, 230);"><span aria-hidden="true">×</span><span class="sr-only">Close</span></button></div><div class="modal-body"><form><div class="form-group"><label for="name">Name</label><input type="text" id="name" class="form-control" placeholder="" value=""><small class="form-text text-muted">Give the state a good name so you can find it later</small></div><div class="form-group"><label for="name">State</label><div class="saved-state-builder"><div class="menu mb-1"><button class="btn btn-secondary">Record</button><button class="btn btn-secondary ml-1">Select all lights</button></div><div class="devices"><div class="selected">Device1</div><div class="selected">Device2</div></div></div></div></form></div><div class="modal-footer"><button type="button" class="btn btn-primary">Use</button></div></div></div></div>
*/
