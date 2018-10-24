import React, { Component } from "react";

class Node extends Component {

    render () {
        const { uuid } = this.props;
        return (
            <div>
                Uuid: {uuid}    
            </div>
        );
    }
}

export default Node;
