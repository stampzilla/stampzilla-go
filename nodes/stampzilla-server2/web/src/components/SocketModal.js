import React, { Component } from 'react';
import Modal from 'react-bootstrap4-modal';

class SocketModal extends Component {
  onChange = () => () => {
    this.props.onChange();
    this.props.onClose();
  }

  render = () => {
    const { visible, onClose } = this.props;

    return (
      <Modal visible={visible} onClickBackdrop={onClose}>
        <div className="modal-header">
          <h5 className="modal-title">Change websocket url</h5>
        </div>
        <div className="modal-body">
          <p>Enemy vessel approaching!</p>
        </div>
        <div className="modal-footer">
          <button type="button" className="btn btn-primary" onClick={this.onChange()}>
              Change
          </button>
        </div>
      </Modal>
    );
  }
}

export default SocketModal;
