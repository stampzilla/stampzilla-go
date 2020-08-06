import React, { Component } from "react";
import { connect } from "react-redux";
import { withRouter } from "react-router";
import { get } from "axios";
import Url from "url";

import { update } from "../ducks/app";
import Link from "./Link";
import SocketModal from "./SocketModal";

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      socketModal: false
    };
  }

  onLogout = () => {
    const { app, server } = this.props;

    const serverUrl = Url.parse(app.get("url"));
    const u = `https://${serverUrl.hostname}:${server.get("tlsPort")}/logout`;

    get(u, { withCredentials: true })
      .then(() => {
        this.props.dispatch(update({ url: "" }));
        this.props.dispatch(update({ url: app.get("url") }));
      })
      .catch(error => {
        console.log("error", error);
      });
  };

  render = () => {
    const { children, connected, dispatch, requests, method } = this.props;
    const { socketModal } = this.state;

    return (
      <>
        <SocketModal
          visible={socketModal}
          onClose={() => this.setState({ socketModal: false })}
          onChange={({ hostname, port }) =>
            dispatch(update({ url: `ws://${hostname}:${port}/ws` }))
          }
        />
        <nav className="main-header navbar navbar-expand bg-white navbar-light border-bottom">
          <ul className="navbar-nav">
            <li className="nav-item">
              <a className="nav-link" data-widget="pushmenu" href="#">
                <i className="fa fa-bars" />
              </a>
            </li>
          </ul>
        </nav>

        <aside className="main-sidebar sidebar-dark-primary elevation-4">
          <a href="" className="brand-link">
            <span className="brand-text font-weight-light">stampzilla-go</span>
          </a>

          <div className="sidebar">
            <nav className="mt-2">
              <ul
                className="nav nav-pills nav-sidebar flex-column"
                data-widget="treeview"
                role="menu"
                data-accordion="false"
              >
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
                  <Link to="/persons" className="nav-link" activeClass="active">
                    <i className="nav-icon fa fa-user" />
                    Persons and users
                  </Link>
                </li>
                <li className="nav-item">
                  <Link to="/nodes" className="nav-link" activeClass="active">
                    <i className="nav-icon fa fa-code" />
                    Nodes
                  </Link>
                </li>
                <li className="nav-item">
                  <Link to="/alerts" className="nav-link" activeClass="active">
                    <i className="nav-icon fa fa-bell" />
                    Alerts
                  </Link>
                </li>
                <li className="nav-item">
                  <Link
                    to="/security"
                    className="nav-link"
                    activeClass="active"
                  >
                    <i className="nav-icon fa fa-shield" />
                    <span>Security</span>
                    {requests && requests.size > 0 && (
                      <div className="badge badge-danger">{requests.size}</div>
                    )}
                  </Link>
                </li>

                <li className="nav-item mt-4">
                  <Link to="/debug" className="nav-link" activeClass="active">
                    <i className="nav-icon fa fa-terminal" />
                    Debug
                  </Link>
                </li>

                {method === "session" && (
                  <li className="nav-item mt-4">
                    <div
                      className="nav-link"
                      onClick={this.onLogout}
                      role="button"
                      tabIndex="-1"
                    >
                      <i className="nav-icon fa fa-sign-out" />
                      Logout
                    </div>
                  </li>
                )}
              </ul>
            </nav>
          </div>
        </aside>

        <div className="content-wrapper">
          {connected === false && (
            <div className="p-4 bg-danger">
              <button
                className="btn btn-secondary float-right"
                type="button"
                style={{ marginTop: "-8px" }}
                onClick={() => this.setState({ socketModal: true })}
              >
                Change socket url
              </button>
              Not connected!
            </div>
          )}
          <div className="content-header">
            <div className="container-fluid">
              <div className="row mb-2">
                <div className="col-sm-6">
                  <h1 className="m-0 text-dark" style={{ display: "none" }}>
                    Debug
                  </h1>
                </div>
              </div>
            </div>
          </div>

          <div className="content">{children}</div>
        </div>

        <aside className="control-sidebar control-sidebar-dark" />
      </>
    );
  };
}

const mapToProps = state => ({
  connected: state.getIn(["connection", "connected"]),
  method: state.getIn(["connection", "method"]),
  requests: state.getIn(["requests", "list"]),
  app: state.getIn(["app"]),
  server: state.getIn(["server"])
});

export default withRouter(connect(mapToProps)(App));
