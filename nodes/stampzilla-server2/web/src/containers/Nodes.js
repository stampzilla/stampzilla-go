import React, { Component } from "react";

import Node from "../components/Node";
import Card from "../components/Card";
import { write } from './Websocket';

class Nodes extends Component {
  
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
          }
        },
      });
    }

    render () {
        return (
            <div>
                <Card
                  title="Send message"
                >
                  <button onClick={this.onClickTestButton()} className="btn btn-primary">Test send message</button>
                </Card>

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
                    <tr>
                      <td>
                        Some Product
                      </td>
                      <td>$13 USD</td>
                      <td>$13 USD</td>
                    </tr>
                    </tbody>
                  </table>
                </Card>
            </div>
        );
    }
}

export default Nodes;
