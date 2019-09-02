import React, { Component } from 'react';
import { Button } from 'reactstrap';
import { connect } from 'react-redux';
import { fromJS } from 'immutable';
import Form from 'react-jsonschema-form';

import {
  ArrayFieldTemplate,
  ConnectedRuleConditions,
  CustomCheckbox,
  ObjectFieldTemplate,
} from '../../components/formComponents';
import { add, save } from '../../ducks/rules';
import Card from '../../components/Card';
import Editor from '../../components/editor';

const schema = fromJS({
  type: 'object',
  required: ['name'],
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    enabled: {
      type: 'boolean',
      title: 'Enabled',
      description: 'Turn on and off this trigger',
    },
    destinations: {
      type: 'array',
      title: 'Destination',
      description: 'Where should the notification be sent?',
      items: {
        type: 'string',
        enum: [],
      },
      uniqueItems: true,
    },
    expression: {
      type: 'string',
      title: 'Expression',
      description:
        'The main expression that describes the state that should activate the trigger',
    },
    for: {
      type: 'string',
      title: 'Delay',
      description:
        'The expression must be fullfilled this amount of time before the trigger is activated',
    },
    conditions: {
      type: 'object',
      title: 'Conditions',
      description:
        'This is the conditions for the trigger to be evaluated. Here you can depend one rule or trigger on an other by selecting if the parent has to be active or not.',
      properties: {},
    },
  },
});
const uiSchema = fromJS({
  expression: {
    'ui:widget': 'Editor',
  },
  conditions: {
    'ui:field': 'ConnectedRuleConditions',
  },
  destinations: {
    'ui:widget': 'checkboxes',
  },
});

const loadFromProps = (props) => {
  const { rules, match } = props;
  const rule = rules.find((n) => n.get('uuid') === match.params.uuid);
  const formData = rule && rule.toJS();

  // if (rule) {
  // formData.actions = formData.actions || [];
  // }
  return {
    formData: {
      ...formData,
      destinations: (formData && formData.destinations) || [],
    },
  };
};

class Trigger extends Component {
  constructor(props) {
    super();

    this.state = {
      isValid: true,
      isModified: false,
      ...loadFromProps(props),
    };
  }

  componentWillReceiveProps(nextProps) {
    const { rules, match } = nextProps;
    if (
      !this.props
      || match.params.uuid !== this.props.match.params.uuid
      || rules !== this.props.rules
    ) {
      this.setState({
        ...loadFromProps(nextProps),
        isModified: false,
      });
    }
  }

  onChange = () => (data) => {
    const { errors, formData } = data;
    this.setState({
      isValid: errors.length === 0,
      formData,
      isModified: true,
    });
  };

  onSubmit = () => ({ formData }) => {
    const { dispatch } = this.props;

    const postData = {
      ...formData,
      type: 'trigger',
      expression: this.form.formElement.querySelector('#editor').value,
    };

    if (formData.uuid) {
      dispatch(save(postData));
    } else {
      dispatch(add(postData));
    }

    this.onBackClick()();
  };

  onBackClick = () => () => {
    const { history } = this.props;
    history.push('/alerts');
  };

  render() {
    const {
      match, devices, state, rules, destinations,
    } = this.props;
    const { isModified } = this.state;

    const params = devices.reduce((acc, dev) => {
      (dev.get('state') || []).forEach((value, key) => {
        acc[`devices["${dev.get('id')}"].${key}`] = value;
      });
      return acc;
    }, {});

    const patchedSchema = schema.toJS();
    if (
      rules.filter((rule) => rule.get('uuid') !== match.params.uuid).size === 0
    ) {
      // Hide the conditions part if there is no other rules
      delete patchedSchema.properties.conditions;
    }

    destinations
      .sort((a, b) => a.get('type').localeCompare(b.get('type')))
      .forEach((dest) => {
        patchedSchema.properties.destinations.items.enum = [
          ...(patchedSchema.properties.destinations.items.enum || []),
          dest.get('uuid'),
        ];
        patchedSchema.properties.destinations.items.enumNames = [
          ...(patchedSchema.properties.destinations.items.enumNames || []),
          `${dest.get('name')} (${dest.get('type')})`,
        ];
      });

    const patchedUiSchema = uiSchema.toJS();
    patchedUiSchema.conditions.current = match.params.uuid;

    return (
      <>
        <div className="row">
          <div className="col-md-12">
            {state.getIn([match.params.uuid, 'error']) && (
              <div className="alert alert-danger">
                {state.getIn([match.params.uuid, 'error'])}
              </div>
            )}
            <Card
              title={(match.params.uuid
                ? 'Edit trigger '
                : 'New trigger'
              ).concat(isModified ? ' (not saved)' : '')}
              bodyClassName="p-0"
            >
              <div className="card-body">
                <Form
                  schema={patchedSchema}
                  uiSchema={patchedUiSchema}
                  showErrorList={false}
                  liveValidate
                  onChange={this.onChange()}
                  formData={this.state.formData}
                  onSubmit={this.onSubmit()}
                  // onError={log('errors')}
                  // disabled={this.props.disabled}
                  // transformErrors={this.props.transformErrors}
                  ObjectFieldTemplate={ObjectFieldTemplate}
                  ArrayFieldTemplate={ArrayFieldTemplate}
                  widgets={{
                    CheckboxWidget: CustomCheckbox,
                    Editor,
                  }}
                  fields={{
                    ConnectedRuleConditions,
                  }}
                  ref={(frm) => {
                    this.form = frm;
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
                <Button color="secondary" onClick={this.onBackClick()}>
                  {'Back'}
                </Button>
                <Button
                  color="primary"
                  disabled={!this.state.isValid || this.props.disabled}
                  onClick={() => this.submitButton.click()}
                  className="float-right"
                >
                  {'Save'}
                </Button>
              </div>
            </Card>

            <pre>
              {Object.keys(params).map((key) => (
                <div key={key}>
                  {key}
:
                  <strong>{JSON.stringify(params[key])}</strong>
                </div>
              ))}
            </pre>
          </div>
        </div>
      </>
    );
  }
}

const mapToProps = (state) => ({
  rules: state.getIn(['rules', 'list']),
  state: state.getIn(['rules', 'state']),
  devices: state.getIn(['devices', 'list']),
  destinations: state.getIn(['destinations', 'list']),
});

export default connect(mapToProps)(Trigger);
