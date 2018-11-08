import React, { Component } from 'react';
import { connect } from 'react-redux';
import { withRouter } from 'react-router';

import { update } from '../ducks/app';
import Link from '../components/Link';
import SocketModal from '../components/SocketModal';

class App extends Component {
  state = {
    socketModal: false,
  }

  render = () => {
    const { children, connected, dispatch } = this.props;
    const { socketModal } = this.state;

    return (
      <React.Fragment>
        <SocketModal
          visible={socketModal}
          onClose={() => this.setState({ socketModal: false })}
          onChange={({ hostname, port }) => dispatch(update({ url: `ws://${hostname}:${port}/ws` }))}
        />
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
                  <Link to="/" className="nav-link" activeClass="active">
                    <i className="nav-icon fa fa-tachometer" />
                      Dashboard
                  </Link>
                </li>
                <li className="nav-item">
                  <Link to="/aut" className="nav-link" activeClass="active">
                    <i className="nav-icon fa fa-magic" />
                      Automation
                  </Link>
                </li>
                <li className="nav-item">
                  <Link to="/nodes" className="nav-link" activeClass="active">
                    <i className="nav-icon fa fa-code" />
                      Nodes
                  </Link>
                </li>
                <li className="nav-item">
                  <Link to="/security" className="nav-link" activeClass="active">
                    <i className="nav-icon fa fa-shield" />
                      Security
                  </Link>
                </li>
                <li className="nav-item mt-4">
                  <Link to="/debug" className="nav-link" activeClass="active">
                    <i className="nav-icon fa fa-terminal" />
                      Debug
                  </Link>
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
                  <h1 className="m-0 text-dark" style={{ display: 'none' }}>Debug</h1>
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

export default withRouter(connect(mapToProps)(App));
