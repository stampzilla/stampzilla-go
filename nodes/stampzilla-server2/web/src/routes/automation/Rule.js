import React, { Component } from 'react';
import {
  Button,
} from 'reactstrap';
import { connect } from 'react-redux';
import Form from 'react-jsonschema-form';

import { write } from '../../components/Websocket';
import Card from '../../components/Card';
import CustomCheckbox from '../../components/CustomCheckbox';

const schema = {
  type: 'object',
  required: [
    'name',
    'config',
  ],
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    enabled: {
      type: 'boolean',
      title: 'Enabled',
      description: 'Turn on and off this rule',
    },
    expression: {
      type: 'string',
      title: 'Expression',
      description: 'The main expression that describes the state that should activate the rule',
    },
  // conditions: {
  // title: 'Conditions',
  // description: 'A list of related rules that should be active or not active to enable this rule',
  // type: "array",
  // items: {
  // type: 'object',
  // properties: {
  // rule: {
  // type: 'string',
  // title: 'Rule',
  // },
  // state: {
  // type: 'boolean',
  // title: 'Active',
  // description: 'Should the rule be active or not?',
  // },
  // },
  // },
  // }
  },
};
const uiSchema = {
  config: {
    'ui:options': {
      rows: 15,
    },
  },
};


class Automation extends Component {
  state = {
    isValid: true,
    formData: {
    },
  }

  onChange = () => (data) => {
    const { errors, formData } = data;
    this.setState({
      isValid: errors.length === 0,
      formData,
    });
  }

  onSubmit = () => ({ formData }) => {
    const { nodes, match } = this.props;
    const node = nodes.find(n => n.get('uuid') === match.params.uuid);

    write({
      type: 'setup-node',
      body: {
        ...node.toJS(),
        ...formData,
        config: JSON.parse(formData.config),
      },
    });
  }

  render() {
    const { match, devices } = this.props;

    const params = devices.reduce((acc, dev) => {
      console.log(dev.toJS());
      dev.get('state').forEach((value, key) => {
        acc[`devices["${dev.get('id')}"].${key}`] = value;
      });
      return acc;
    }, {});

    return (
      <React.Fragment>
        <div className="row">
          <div className="col-md-12">
            <Card
              title={match.params.uuid ? 'Edit rule ' : 'New rule'}
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
                  // onError={log('errors')}
                  // disabled={this.props.disabled}
                  // transformErrors={this.props.transformErrors}
                  widgets={{
                    CheckboxWidget: CustomCheckbox,
                  }}
                >
                  <button ref={(btn) => { this.submitButton = btn; }} style={{ display: 'none' }} type="submit" />
                </Form>


                <pre>
                  {Object.keys(params).map(key => (
                    <div>{key}: {params[key]}</div>
                  ))}
                </pre>
              </div>
              <div className="card-footer">
                <Button color="primary" disabled={!this.state.isValid || this.props.disabled} onClick={() => this.submitButton.click()}>
                  {'Save'}
                </Button>
              </div>
            </Card>
          </div>
        </div>
      </React.Fragment>
    );
  }
}

const mapToProps = state => ({
  rules: state.getIn(['rules', 'list']),
  devices: state.getIn(['devices', 'list']),
});

export default connect(mapToProps)(Automation);
