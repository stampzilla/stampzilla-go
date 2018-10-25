import React, { Component } from "react";
import Node from "../components/Node";

import { write } from './Websocket'; 

class Nodes extends Component {
    constructor() {
        super();

        this.state = {
            nodes: ['a', 'b', 'c', 'd']
        };
    }
  
    onClickTestButton = () => () => {
      write({
        type: 'update-node',
        message: {
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
        const { nodes } = this.state;

        return (
            <div>
                <button onClick={this.onClickTestButton()}>Test send message</button>
                {nodes.map(node => <Node key={node} uuid={node} />)}    
            </div>
        );
    }
}

export default Nodes;
