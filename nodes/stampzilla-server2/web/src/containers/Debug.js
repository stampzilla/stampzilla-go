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
                      <th>Address</th>
                      <th>Type</th>
                    </tr>
                  </thead>
                  <tbody>
                    {connections.map(message => (
                      <tr>
                        <td>{message.remoteAddr}</td>
                        <td>{message.type}</td>
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
