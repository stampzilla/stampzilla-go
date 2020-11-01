import React, { Component } from 'react';
import { Button } from 'reactstrap';
import { connect } from 'react-redux';
import Form from 'react-jsonschema-form';

import { request } from '../../components/Websocket';
import Card from '../../components/Card';
import CustomCheckbox from '../../components/CustomCheckbox';

const schema = {
  type: 'object',
  required: ['server', 'instance'],
  properties: {
    server: {
      type: 'string',
      title: 'Server',
    },
    instance: {
      type: 'string',
      title: 'Instance',
    },
  },
};
const uiSchema = {};

class Person extends Component {
  state = {
    isValid: true,
    formData: {},
  };

  componentDidMount = () => {
    this.componentWillReceiveProps(this.props);
  };

  componentWillReceiveProps = (props) => {
    const { config } = props;
    this.setState({
      formData: config.toJS(),
    });
  };

  onChange = () => (data) => {
    const { errors, formData } = data;
    this.setState({
      isValid: errors.length === 0,
      formData,
    });
  };

  onValidate = (formData, errors) => {
    if (formData.new_password !== formData.repeat_password) {
      errors.repeat_password.addError("Passwords don't match");
    }
    return errors;
  };

  onSubmit = () => async ({ formData }) => {
    const { config } = this.props;

    const postData = {
      ...config.toJS(),
      ...formData,
      enable: true,
    };

    // write({
    // type: 'update-person',
    // body: {
    // ...person.toJS(),
    // ...formData,
    // },
    // });
    await request({ type: 'cloud-connect', body: postData });
  };

  onDisconnect = async () => {
    await request({ type: 'cloud-disconnect' });
  };

  render() {
    const { config, state } = this.props;

    return (
      <div className="content">
        <div className="row">
          {config.get('enable') && (
            <div className="col-md-12">
              <Card title="Cloud connection status" bodyClassName="p-0">
                <div className="card-body">
                  {state.get('error') && (
                    <div className="alert alert-danger">
                      <strong>Connection error:</strong>
                      {' '}
                      {state.get('error')}
                    </div>
                  )}
                  <table className="table table-striped table-valign-middle">
                    <tbody>
                      <tr>
                        <th>Instance name</th>
                        <td>{config.get('instance')}</td>
                      </tr>
                      <tr>
                        <th>Server</th>
                        <td>{config.get('server')}</td>
                      </tr>
                      <tr>
                        <th>Connected</th>
                        <td>
                          {state.get('connected') ? (
                            <div className="badge badge-success">Yes</div>
                          ) : (
                            <div className="badge badge-danger">No</div>
                          )}
                        </td>
                      </tr>
                      <tr>
                        <th>Using TLS encryption</th>
                        <td>
                          {state.get('secure') ? (
                            <div className="badge badge-success">Yes</div>
                          ) : (
                            <div className="badge badge-danger">No</div>
                          )}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <div className="card-footer">
                  <Button
                    color="warning"
                    onClick={this.onDisconnect}
                    className="float-right"
                  >
                    Disconnect
                  </Button>
                </div>
              </Card>
            </div>
          )}

          {!config.get('enable') && (
            <div className="col-md-12">
              <Card
                title="Activate a cloud proxy connection"
                bodyClassName="p-0"
              >
                <div className="card-body">
                  <Form
                    schema={schema}
                    uiSchema={uiSchema}
                    showErrorList={false}
                    liveValidate
                    onChange={this.onChange()}
                    formData={this.state.formData}
                    onSubmit={this.onSubmit()}
                    validate={this.onValidate}
                    // onError={log('errors')}
                    // disabled={this.props.disabled}
                    // transformErrors={this.props.transformErrors}
                    widgets={{
                      CheckboxWidget: CustomCheckbox,
                    }}
                  >
                    <button
                      ref={(btn) => {
                        this.submitButton = btn;
                      }}
                      style={{ display: 'none' }}
                      type="submit"
                    />
                  </Form>
                </div>
                <div className="card-footer">
                  <Button
                    color="primary"
                    disabled={!this.state.isValid || this.props.disabled}
                    onClick={() => this.submitButton.click()}
                    className="float-right"
                  >
                    Connect
                  </Button>
                </div>
              </Card>
            </div>
          )}
        </div>
      </div>
    );
  }
}

const mapToProps = (state) => ({
  config: state.getIn(['cloud', 'config']),
  state: state.getIn(['cloud', 'state']),
});

export default connect(mapToProps)(Person);
