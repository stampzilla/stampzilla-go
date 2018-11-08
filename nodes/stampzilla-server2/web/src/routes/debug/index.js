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
          state: {
            connected: true,
            writeTest: false,
          },
          writemap: {
            connected: false,
            writeTest: true,
          },
          config: {
            background: '#fff',
          },
        },
      });
    }

    render() {
      const { messages, connections } = this.props;

      return (
        <React.Fragment>
          <div className="row">
            <div className="col-md-6">
              <Card
                title="Send message"
              >
                <button onClick={this.onClickTestButton()} className="btn btn-primary">Test send message</button>
              </Card>

            </div>
            <div className="col-md-6">
              <Card
                title="Server connections"
                bodyClassName="p-0"
              >
                <table className="table table-striped table-valign-middle">
                  <thead>
                    <tr>
                      <th />
                      <th>Identity</th>
                      <th>Address</th>
                      <th>Type</th>
                    </tr>
                  </thead>
                  <tbody>
                    {connections
                      .sort((a, b) => {
                        if (a.getIn(['attributes', 'secure']) !== b.getIn(['attributes', 'secure'])) {
                          return a.getIn(['attributes', 'secure']) ? -1 : 1;
                        }

                        const t = b.get('type').localeCompare(a.get('type'));
                        if (t !== 0) {
                          return t;
                        }

                        return a.getIn(['attributes', 'identity']).localeCompare(b.getIn(['attributes', 'identity']));
                      })
                      .map(c => (
                        <tr key={c.get('id')}>
                          <td>
                            {c.getIn(['attributes', 'secure']) &&
                            <React.Fragment>
                              {c.getIn(['attributes', 'identity']) &&
                                <i className="nav-icon fa fa-lock text-success" />
                              }
                              {!c.getIn(['attributes', 'identity']) &&
                                <i className="nav-icon fa fa-lock" />
                              }
                            </React.Fragment>
                            }
                          </td>
                          <td>{c.getIn(['attributes', 'identity'])}</td>
                          <td>{c.get('remoteAddr')}</td>
                          <td>{c.get('type')}</td>
                        </tr>
                    )).toArray()}
                  </tbody>
                </table>
              </Card>
            </div>
          </div>
          <Card
            title="Command bus"
            bodyClassName="p-0"
          >
            <table className="table table-striped table-valign-middle">
              <thead>
                <tr>
                  <th>Node</th>
                  <th>Type</th>
                  <th>Body</th>
                </tr>
              </thead>
              <tbody>
                {messages.reverse().map(message => (
                  <tr>
                    <td />
                    <td>{message.type}</td>
                    <td>{JSON.stringify(message.body)}</td>
                  </tr>
                    ))}
              </tbody>
            </table>
          </Card>
        </React.Fragment>
      );
    }
}

const mapToProps = state => ({
  messages: state.getIn(['connection', 'messages']),
  connections: state.getIn(['connections', 'list']),
});

export default connect(mapToProps)(Debug);
