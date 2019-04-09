import React, { Component } from 'react';
import { connect } from 'react-redux';

import Card from '../../components/Card';

const toStatusBadge = (n, state) => {
  if (n.get('enabled')) {
    if (!state) {
      return <div className="badge badge-info">Unknown</div>;
    }

    if (state.get('error')) {
      return <div className="badge badge-danger">Failed</div>;
    }
    if (state.get('active')) {
      return <div className="badge badge-success">Active</div>;
    }
    if (state.get('pending')) {
      return <div className="badge badge-warning">Pending</div>;
    }

    return <div className="badge badge-secondary">Standby</div>;
  }

  return <div className="badge badge-secondary">Disabled</div>;
};

class Automation extends Component {
  onClickNode = uuid => () => {
    const { history } = this.props;
    history.push(`/aut/${uuid}`);
  };

  render() {
    const {
      rules, rulesState, destinations, senders,
    } = this.props;

    return (
      <React.Fragment>
        <div className="row">
          <div className="col-md-12">
            <Card
              title={(
                <React.Fragment>
                  Triggers
                  {' '}
                  <small>
                    (Rules that will create a message and sends it to a
                    destination)
                  </small>
                </React.Fragment>
)}
              bodyClassName="p-0"
              toolbar={[
                {
                  icon: 'fa fa-plus',
                  className: 'btn-secondary',
                  onClick: this.onClickNode('triggers/create'),
                },
              ]}
            >
              <table className="table table-striped table-valign-middle">
                <thead>
                  <tr>
                    <th style={{ width: 1 }}>Status</th>
                    <th>Name</th>
                  </tr>
                </thead>
                <tbody>
                  {rules
                    && rules
                      .filter(() => false) // TODO:
                      .sort((a, b) => a.get('name').localeCompare(b.get('name')))
                      .map(n => (
                        <tr
                          key={n.get('uuid')}
                          style={{ cursor: 'pointer' }}
                          onClick={this.onClickNode(`rule/${n.get('uuid')}`)}
                        >
                          <td className="text-center">
                            {toStatusBadge(n, rulesState.get(n.get('uuid')))}
                          </td>
                          <td>{n.get('name') || n.get('uuid')}</td>
                        </tr>
                      ))
                      .valueSeq()
                      .toArray()}
                </tbody>
              </table>
            </Card>
            <div
              className="text-center pb-3"
              style={{ fontSize: '3em', color: '#ccc' }}
            >
              <span className="fa fa-arrow-down" />
            </div>

            <Card
              title={(
                <React.Fragment>
                  Destinations
                  {' '}
                  <small>(Groups of addresses and destinations)</small>
                </React.Fragment>
)}
              bodyClassName="p-0"
              toolbar={[
                {
                  icon: 'fa fa-plus',
                  className: 'btn-secondary',
                  onClick: this.onClickNode('destinations/create'),
                },
              ]}
            >
              <table className="table table-striped table-valign-middle">
                <thead>
                  <tr>
                    <th>Name</th>
                  </tr>
                </thead>
                <tbody>
                  {destinations
                    && destinations
                      .sort((a, b) => a.get('name').localeCompare(b.get('name')))
                      .map(n => (
                        <tr
                          key={n.get('uuid')}
                          style={{ cursor: 'pointer' }}
                          onClick={this.onClickNode(
                            `destinations/${n.get('uuid')}`,
                          )}
                        >
                          <td>{n.get('name') || n.get('uuid')}</td>
                        </tr>
                      ))
                      .valueSeq()
                      .toArray()}
                </tbody>
              </table>
            </Card>
            <div
              className="text-center pb-3"
              style={{ fontSize: '3em', color: '#ccc' }}
            >
              <span className="fa fa-arrow-down" />
            </div>
            <Card
              title={(
                <React.Fragment>
                  Senders
                  {' '}
                  <small>(services that can deliver messages)</small>
                </React.Fragment>
)}
              bodyClassName="p-0"
              toolbar={[
                {
                  icon: 'fa fa-plus',
                  className: 'btn-secondary',
                  onClick: this.onClickNode('senders/create'),
                },
              ]}
            >
              <table className="table table-striped table-valign-middle">
                <thead>
                  <tr>
                    <th>Name</th>
                  </tr>
                </thead>
                <tbody>
                  {senders
                    && senders
                      .sort((a, b) => a.get('name').localeCompare(b.get('name')))
                      .map(n => (
                        <tr
                          key={n.get('uuid')}
                          style={{ cursor: 'pointer' }}
                          onClick={this.onClickNode(
                            `schedule/${n.get('uuid')}`,
                          )}
                        >
                          <td>{n.get('name')}</td>
                        </tr>
                      ))
                      .valueSeq()
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
  rules: state.getIn(['rules', 'list']),
  rulesState: state.getIn(['rules', 'state']),
  destinations: state.getIn(['destinations', 'list']),
  senders: state.getIn(['senders', 'list']),
});

export default connect(mapToProps)(Automation);
