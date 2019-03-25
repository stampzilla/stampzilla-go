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
      rules, schedules, rulesState, schedulesState,
    } = this.props;

    return (
      <React.Fragment>
        <div className="row">
          <div className="col-md-12">
            <Card
              title="Rules"
              bodyClassName="p-0"
              toolbar={[
                {
                  icon: 'fa fa-plus',
                  className: 'btn-secondary',
                  onClick: this.onClickNode('rule/create'),
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

            <Card
              title="Schedules"
              bodyClassName="p-0"
              toolbar={[
                {
                  icon: 'fa fa-plus',
                  className: 'btn-secondary',
                  onClick: this.onClickNode('schedule/create'),
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
                  {schedules
                    && schedules
                      .sort((a, b) => a.get('name').localeCompare(b.get('name')))
                      .map(n => (
                        <tr
                          key={n.get('uuid')}
                          style={{ cursor: 'pointer' }}
                          onClick={this.onClickNode(
                            `schedule/${n.get('uuid')}`,
                          )}
                        >
                          <td className="text-center">
                            {toStatusBadge(
                              n,
                              schedulesState.get(n.get('uuid')),
                            )}
                          </td>
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
  schedules: state.getIn(['schedules', 'list']),
  schedulesState: state.getIn(['schedules', 'state']),
});

export default connect(mapToProps)(Automation);
