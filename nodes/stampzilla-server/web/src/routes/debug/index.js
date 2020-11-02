import React, { Component } from 'react';
import { connect } from 'react-redux';

import { write } from '../../components/Websocket';
import Card from '../../components/Card';

class Debug extends Component {
  onClickTestButton = () => () => {
    write({
      type: 'update-node',
      body: {
        uuid: '123',
        version: '1',
        name: 'web client',
        config: 'trasig json;',
      },
    });
  };

  render() {
    const {
      messages, connections, nodes, persons,
    } = this.props;

    return (
      <>
        <div className="row">
          <div className="col-md-12">
            <Card title="Active connections" bodyClassName="p-0">
              <table className="table table-striped table-valign-middle">
                <thead>
                  <tr>
                    <th />
                    <th>Identity</th>
                    <th>Address</th>
                    <th>Type</th>
                    <th>Subscriptions</th>
                  </tr>
                </thead>
                <tbody>
                  {connections
                    .sort((a, b) => {
                      if (
                        a.getIn(['attributes', 'secure'])
                        !== b.getIn(['attributes', 'secure'])
                      ) {
                        return a.getIn(['attributes', 'secure']) ? -1 : 1;
                      }

                      const t = b.get('type').localeCompare(a.get('type'));
                      if (t !== 0) {
                        return t;
                      }

                      return (
                        a.getIn(['attributes', 'identity']) || ''
                      ).localeCompare(b.getIn(['attributes', 'identity']));
                    })
                    .map((c) => (
                      <tr key={c.getIn(['attributes', 'ID'])}>
                        <td>
                          {c.getIn(['attributes', 'secure']) && (
                            <>
                              {c.getIn(['attributes', 'identity']) && (
                                <i className="nav-icon fa fa-lock text-success" />
                              )}
                              {!c.getIn(['attributes', 'identity']) && (
                                <i className="nav-icon fa fa-lock" />
                              )}
                            </>
                          )}
                        </td>
                        <td>
                          {persons.getIn([
                            c.getIn(['attributes', 'identity']),
                            'name',
                          ])
                            || nodes.getIn([
                              c.getIn(['attributes', 'identity']),
                              'name',
                            ])
                            || c.getIn(['attributes', 'ID'])}
                        </td>
                        <td>{c.get('remoteAddr')}</td>
                        <td>{c.get('type')}</td>
                        <td>
                          {c.getIn(['attributes', 'subscriptions'])
                            && c
                              .getIn(['attributes', 'subscriptions'])
                              .sort()
                              .join(', ')}
                        </td>
                      </tr>
                    ))
                    .valueSeq()
                    .toArray()}
                </tbody>
              </table>
            </Card>
          </div>
        </div>
        <Card title="Received messages" bodyClassName="p-0">
          <table className="table table-striped table-valign-middle">
            <thead>
              <tr>
                <th>Count</th>
                <th>Type</th>
                <th>Body</th>
              </tr>
            </thead>
            <tbody>
              {messages.reverse().map((message) => (
                <tr key={message.id}>
                  <td>{message.id}</td>
                  <td>{message.type}</td>
                  <td>{JSON.stringify(message.body)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </>
    );
  }
}

const mapToProps = (state) => ({
  messages: state.getIn(['connection', 'messages']),
  connections: state.getIn(['connections', 'list']),
  nodes: state.getIn(['nodes', 'list']),
  persons: state.getIn(['persons', 'list']),
});

export default connect(mapToProps)(Debug);
