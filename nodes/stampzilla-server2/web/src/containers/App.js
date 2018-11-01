import ReconnectableWebSocket from 'reconnectable-websocket'
import React, { Component } from "react";
import { connect } from 'react-redux';
import Modal from 'react-bootstrap4-modal';


class App extends Component {
  state = {
    socketModal: false,
  }

  render = () => {
    const { children, connected } = this.props;
    const { socketModal } = this.state;

    return (
      <div class="wrapper">
        <Modal visible={socketModal} onClickBackdrop={() => this.setState({ socketModal: false })}>
          <div className="modal-header">
            <h5 className="modal-title">Change websocket url</h5>
          </div>
		  <div className="modal-body">
		    <p>Enemy vessel approaching!</p>
		  </div>
		  <div className="modal-footer">
			<button type="button" className="btn btn-primary" onClick={this.onFirePhasers}>
			  Change
			</button>
		  </div>
        </Modal>
        <nav class="main-header navbar navbar-expand bg-white navbar-light border-bottom">

          <ul class="navbar-nav">
            <li class="nav-item">
              <a class="nav-link" data-widget="pushmenu" href="#"><i class="fa fa-bars"></i></a>
            </li>
          </ul>
        </nav>


        <aside class="main-sidebar sidebar-dark-primary elevation-4">
          <a href="" class="brand-link">
            <span class="brand-text font-weight-light">stampzilla-go</span>
          </a>

          <div class="sidebar">
            <nav class="mt-2">
              <ul class="nav nav-pills nav-sidebar flex-column" data-widget="treeview" role="menu" data-accordion="false">
                <li class="nav-item">
                  <a href="#" class="nav-link active">
                    <i class="nav-icon fa fa-terminal"></i>
                      Debug
                  </a>
                </li>
              </ul>
            </nav>
          </div>
        </aside>

        <div class="content-wrapper">
          {!connected && 
          <div class="p-4 bg-danger" >
            <button className="btn btn-secondary float-right" style={{ marginTop: "-8px"}} onClick={() => this.setState({ socketModal: true })}>Change socket url</button>
            Not connected!
          </div>
          }
          <div class="content-header">
            <div class="container-fluid">
              <div class="row mb-2">
                <div class="col-sm-6">
                  <h1 class="m-0 text-dark">Debug</h1>
                </div>
              </div>
            </div>
          </div>

          <div class="content">
            {children}
          </div>
        </div>

        <aside class="control-sidebar control-sidebar-dark">
        </aside>
      </div>
    );
  };
}

const mapToProps = (state) => ({
  connected: state.getIn(['connection', 'connected']),
});

export default connect(mapToProps)(App);
