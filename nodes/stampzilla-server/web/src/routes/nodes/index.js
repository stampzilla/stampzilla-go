import React, { Component } from 'react';
import { connect } from 'react-redux';

import { write } from '../../components/Websocket';
import Card from '../../components/Card';

class Nodes extends Component {
  onClickTestButton = () => () => {
    write({
      type: 'update-node',
      body: {},
    });
  };

  onClickNode = uuid => () => {
    const { history } = this.props;
    history.push(`/nodes/${uuid}`);
  };

  render() {
    const { nodes, connections } = this.props;

    return (
      <React.Fragment>
        <div className="row">
          <div className="col-md-12">
            <Card title="Nodes" bodyClassName="p-0">
              <table className="table table-striped table-valign-middle">
                <thead>
                  <tr>
                    <th>Connected</th>
                    <th>Name</th>
                    <th>Identity</th>
                  </tr>
                </thead>
                <tbody>
                  {nodes
                    .sort((a, b) =>
                      (a.get('name') || '').localeCompare(b.get('name')))
                    .map(n => (
                      <tr
                        key={n.get('uuid')}
                        style={{ cursor: 'pointer' }}
                        onClick={this.onClickNode(n.get('uuid'))}
                      >
                        <td>
                          {connections.getIn([
                            n.get('uuid'),
                            'attributes',
                            'secure',
                          ]) && (
                            <React.Fragment>
                              {connections.getIn([
                                n.get('uuid'),
                                'attributes',
                                'identity',
                              ]) && (
                                <i
                                  className="nav-icon fa fa-lock text-success"
                                  title="Securly connected"
                                />
                              )}
                              {!connections.getIn([
                                n.get('uuid'),
                                'attributes',
                                'identity',
                              ]) && (
                                <i
                                  className="nav-icon fa fa-lock"
                                  title="Missing identity"
                                />
                              )}
                            </React.Fragment>
                          )}
                          {!connections.has(n.get('uuid')) && (
                            <i
                              className="nav-icon fa fa-exclamation-triangle text-warning"
                              title="Not connected"
                            />
                          )}
                        </td>
                        <td>{n.get('name')}</td>
                        <td>
                          <small>{n.get('uuid')}</small>
                        </td>
                      </tr>
                    ))
                    .toArray()}
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
  nodes: state.getIn(['nodes', 'list']),
  connections: state.getIn(['connections', 'list']),
});

export default connect(mapToProps)(Nodes);
