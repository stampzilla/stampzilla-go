import React, { Component } from 'react';
import { connect } from 'react-redux';
import Modal from 'react-bootstrap4-modal';

import { update } from '../ducks/app';

class App extends Component {
  state = {
    socketModal: false,
  }

  onSocketModalClick = () => () => {
    this.props.dispatch(update({
      url: 'asd',
    }));
  }

  render = () => {
    const { children, connected } = this.props;
    const { socketModal } = this.state;

    return (
      <React.Fragment>
        <Modal visible={socketModal} onClickBackdrop={() => this.setState({ socketModal: false })}>
          <div className="modal-header">
            <h5 className="modal-title">Change websocket url</h5>
          </div>
          <div className="modal-body">
            <p>Enemy vessel approaching!</p>
          </div>
          <div className="modal-footer">
            <button type="button" className="btn btn-primary" onClick={this.onSocketModalClick()}>
              Change
            </button>
          </div>
        </Modal>
        <nav className="main-header navbar navbar-expand bg-white navbar-light border-bottom">

          <ul className="navbar-nav">
            <li className="nav-item">
              <a className="nav-link" data-widget="pushmenu" href="#"><i className="fa fa-bars" /></a>
            </li>
          </ul>
        </nav>


        <aside className="main-sidebar sidebar-dark-primary elevation-4">
          <a href="" className="brand-link">
            <span className="brand-text font-weight-light">stampzilla-go</span>
          </a>

          <div className="sidebar">
            <nav className="mt-2">
              <ul className="nav nav-pills nav-sidebar flex-column" data-widget="treeview" role="menu" data-accordion="false">
                <li className="nav-item">
                  <a href="#" className="nav-link active">
                    <i className="nav-icon fa fa-terminal" />
                      Debug
                  </a>
                </li>
              </ul>
            </nav>
          </div>
        </aside>

        <div className="content-wrapper">
          {connected === false &&
          <div className="p-4 bg-danger" >
            <button className="btn btn-secondary float-right" style={{ marginTop: '-8px' }} onClick={() => this.setState({ socketModal: true })}>Change socket url</button>
            Not connected!
          </div>
          }
          <div className="content-header">
            <div className="container-fluid">
              <div className="row mb-2">
                <div className="col-sm-6">
                  <h1 className="m-0 text-dark">Debug</h1>
                </div>
              </div>
            </div>
          </div>

          <div className="content">
            {children}
          </div>
        </div>

        <aside className="control-sidebar control-sidebar-dark" />
      </React.Fragment>
    );
  };
}

const mapToProps = state => ({
  connected: state.getIn(['connection', 'connected']),
});

export default connect(mapToProps)(App);
