import React, { Component } from 'react';
import { connect } from 'react-redux';

import { write } from '../../components/Websocket';
import Card from '../../components/Card';

class Persons extends Component {
  onClickTestButton = () => () => {
    write({
      type: 'update-node',
      body: {},
    });
  };

  onClickPerson = (uuid) => () => {
    const { history } = this.props;
    history.push(`/persons/${uuid}`);
  };

  render() {
    const { persons } = this.props;

    return (
      <>
        <div className="row">
          <div className="col-md-12">
            <Card title="Persons" bodyClassName="p-0">
              <table className="table table-striped table-valign-middle">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Identity</th>
                  </tr>
                </thead>
                <tbody>
                  {persons
                    .sort((a, b) => (a.get('name') || '').localeCompare(b.get('name')))
                    .map((n) => (
                      <tr
                        key={n.get('uuid')}
                        style={{ cursor: 'pointer' }}
                        onClick={this.onClickPerson(n.get('uuid'))}
                      >
                        <td>{n.get('name')}</td>
                        <td>
                          <small>{n.get('uuid')}</small>
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
      </>
    );
  }
}

const mapToProps = (state) => ({
  persons: state.getIn(['persons', 'list']),
});

export default connect(mapToProps)(Persons);
