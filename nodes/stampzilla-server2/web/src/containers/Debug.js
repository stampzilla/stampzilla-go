import React, { Component } from 'react';
import { connect } from 'react-redux';

import { write } from './Websocket';
import Card from '../components/Card';

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
                        if (a.attributes.secure !== b.attributes.secure) {
                          return a.attributes.secure ? -1 : 1;
                        }

                        const t = a.type.localeCompare(b.type);
                        if (t !== 0) {
                          return -t;
                        }

                        if (typeof a.attributes.identity === 'undefined') {
                          return -1;
                        }

                        if (typeof b.attributes.identity === 'undefined') {
                          return 1;
                        }
                        return a.attributes.identity.localeCompare(b.attributes.identity);
                      })
                      .map(c => (
                        <tr key={c.attributes.ID}>
                          <td>
                            {c.attributes.secure &&
                            <React.Fragment>
                              {c.attributes.identity &&
                                <i className="nav-icon fa fa-lock text-success" />
                              }
                              {!c.attributes.identity &&
                                <i className="nav-icon fa fa-lock" />
                              }
                            </React.Fragment>
                            }
                          </td>
                          <td>{c.attributes.identity}</td>
                          <td>{c.remoteAddr}</td>
                          <td>{c.type}</td>
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
