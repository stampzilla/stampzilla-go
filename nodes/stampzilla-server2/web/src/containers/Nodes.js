import React, { Component } from "react";
import Node from "../components/Node";

class Nodes extends Component {
    constructor() {
        super();

        this.state = {
            nodes: ['a', 'b', 'c', 'd']
        };
    }

    render () {
        const { nodes } = this.state;

        return (
            <div>
                {nodes.map(node => <Node key={node} uuid={node} />)}    
            </div>
        );
    }
}

export default Nodes;
