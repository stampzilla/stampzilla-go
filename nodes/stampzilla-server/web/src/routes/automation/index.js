import React, { Component } from 'react';
import { connect } from 'react-redux';

import Card from '../../components/Card';

class Automation extends Component {
    onClickNode = uuid => () => {
      const { history } = this.props;
      history.push(`/aut/${uuid}`);
    }

    render() {
      const { rules, schedules } = this.props;

      return (
        <React.Fragment>
          <div className="row">
            <div className="col-md-12">
              <Card
                title="Rules"
                bodyClassName="p-0"
                toolbar={[{
                  icon: 'fa fa-plus',
                  className: 'btn-secondary',
                  onClick: this.onClickNode('rule/create'),
                }]}
              >
                <table className="table table-striped table-valign-middle">
                  <thead>
                    <tr>
                      <th>Identity</th>
                      <th>Name</th>
                    </tr>
                  </thead>
                  <tbody>
                    {rules && rules
                      .map(n => (
                        <tr key={n.get('uuid')} style={{ cursor: 'pointer' }} onClick={this.onClickNode(`rule/${n.get('uuid')}`)}>
                          <td>{n.get('uuid')}</td>
                          <td>{n.get('name')}</td>
                        </tr>
                    )).toArray()}
                  </tbody>
                </table>
              </Card>

              <Card
                title="Schedules"
                bodyClassName="p-0"
                toolbar={[{
                  icon: 'fa fa-plus',
                  className: 'btn-secondary',
                  onClick: this.onClickNode('schedule/create'),
                }]}
              >
                <table className="table table-striped table-valign-middle">
                  <thead>
                    <tr>
                      <th>Identity</th>
                      <th>Name</th>
                    </tr>
                  </thead>
                  <tbody>
                    {schedules && schedules
                      .map(n => (
                        <tr key={n.get('uuid')} style={{ cursor: 'pointer' }} onClick={this.onClickNode(`schedule/${n.get('uuid')}`)}>
                          <td>{n.get('uuid')}</td>
                          <td>{n.get('name')}</td>
                        </tr>
                    )).toArray()}
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
  schedules: state.getIn(['schedules', 'list']),
});

export default connect(mapToProps)(Automation);