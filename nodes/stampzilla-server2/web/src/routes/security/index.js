import React, { Component } from 'react';
import { connect } from 'react-redux';
import Moment from 'react-moment';

import { write } from '../../components/Websocket';
import Card from '../../components/Card';

class Security extends Component {
    onClickNode = uuid => () => {
      const { history } = this.props;
      history.push(`/security/${uuid}`);
    }

  onClickAccept = (connection) => () => {
    write({
      type: 'accept-request',
      body: connection,
    });
  }

    render() {
      const { certificates, requests, connections } = this.props;

      return (
        <React.Fragment>
          {requests.size > 0 &&
          <div className="row">
            <div className="col-md-12">
              <Card
                title="New node requests"
                bodyClassName="p-0"
              >
                <table className="table table-striped table-valign-middle">
                  <thead>
                    <tr>
                      <th>Requested identity</th>
                      <th>Type</th>
                      <th>Connecting from</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {requests
                      .map(n => (
                        <tr key={n.get('connection')}>
                          <td>{n.get('identity')}</td>
                          <td>{connections.getIn([n.get('connection'), 'type'])} ({n.get('type')})</td>
                          <td>{connections.getIn([n.get('connection'), 'remoteAddr'])}</td>
                          <td className="text-right">
                            <button 
                              className="btn btn-success" 
                              onClick={this.onClickAccept(n.get('connection'))}
                            >Accept</button></td>
                        </tr>
                    )).toArray()}
                  </tbody>
                </table>
              </Card>
            </div>
          </div>
          }

          <div className="row">
            <div className="col-md-12">
              <Card
                title="Certificates"
                bodyClassName="p-0"
              >
                <table className="table table-striped table-valign-middle">
                  <thead>
                    <tr>
                      <th>Common name</th>
                      <th>Type</th>
                      <th>Issued</th>
                      <th>Fingerprint (sha1)</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {certificates
                      .map(n => (
                        <tr key={n.get('serial')} style={{ cursor: 'pointer' }} onClick={this.onClickNode(n.get('serial'))}>
                          <td>{n.get('commonName')}</td>
                          <td>{n.get('usage').sort().join(', ')}</td>
                          <td><Moment fromNow withTitle>{n.get('issued')}</Moment></td>
                          <td>{n.getIn(['fingerprints', 'sha1'])}</td>
                          <td className="text-right"><button className="btn btn-danger" disabled>Revoke</button></td>
                        </tr>
                    )).toArray()}
                  </tbody>
                </table>
              </Card>
            </div>
          </div>
        </React.Fragment>
      );
    }
}

const mapToProps = state => ({
  certificates: state.getIn(['certificates', 'list']),
  requests: state.getIn(['requests', 'list']),
  connections: state.getIn(['connections', 'list']),
});

export default connect(mapToProps)(Security);
