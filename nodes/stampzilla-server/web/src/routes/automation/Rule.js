import React, { Component } from 'react';
import { Button } from 'reactstrap';
import { connect } from 'react-redux';
import Form from 'react-jsonschema-form';
import { fromJS } from 'immutable';

import Editor from './components/editor';
import { add, save } from '../../ducks/rules';
import Card from '../../components/Card';
import SavedStateWidget from './components/SavedStatePicker';
import {
  ArrayFieldTemplate,
  CustomCheckbox,
  ObjectFieldTemplate,
  ConnectedRuleConditions,
} from './components/formComponents';

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
      description: 'Turn on and off this rule',
    },
    expression: {
      type: 'string',
      title: 'Expression',
      description:
        'The main expression that describes the state that should activate the rule',
    },
    for: {
      type: 'string',
      title: 'Delay',
      description:
        'The expression must be fullfilled this amount of time before the actions are carried out',
    },
    actions: {
      type: 'array',
      title: 'Actions',
      items: {
        type: 'string',
      },
    },
    conditions: {
      type: 'object',
      title: 'Conditions',
      description:
        'This is the conditions for the rule to be evaluated. Here you can depend one rule on an other by selecting if the parent rule has to be active or not.',
      properties: {},
    },
  },
});
const uiSchema = fromJS({
  expression: {
    'ui:widget': 'Editor',
  },
  actions: {
    items: {
      'ui:widget': 'SavedStateWidget',
    },
  },
  conditions: {
    'ui:field': 'ConnectedRuleConditions',
  },
});

const loadFromProps = (props) => {
  const { rules, match } = props;
  const rule = rules.find(n => n.get('uuid') === match.params.uuid);
  const formData = rule && rule.toJS();

  if (rule) {
    formData.actions = formData.actions || [];
  }
  return { formData };
};

class Automation extends Component {
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
      expression: this.form.formElement.querySelector('#editor').value,
    };

    if (formData.uuid) {
      dispatch(save(postData));
    } else {
      dispatch(add(postData));
    }
  };

  onBackClick = () => () => {
    const { history } = this.props;
    history.push('/aut');
  };

  render() {
    const {
      match, devices, state, rules,
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
      rules.filter(rule => rule.get('uuid') !== match.params.uuid).size === 0
    ) {
      // Hide the conditions part if there is no other rules
      delete patchedSchema.properties.conditions;
    }

    const patchedUiSchema = uiSchema.toJS();
    patchedUiSchema.conditions.current = match.params.uuid;

    return (
      <React.Fragment>
        <div className="row">
          <div className="col-md-12">
            {state.getIn([match.params.uuid, 'error']) && (
              <div className="alert alert-danger">
                {state.getIn([match.params.uuid, 'error'])}
              </div>
            )}
            <Card
              title={match.params.uuid ? 'Edit rule ' : 'New rule'}
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
                    SavedStateWidget,
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
              {Object.keys(params).map(key => (
                <div key={key}>
                  {key}
:
                  <strong>{JSON.stringify(params[key])}</strong>
                </div>
              ))}
            </pre>
          </div>
        </div>
      </React.Fragment>
    );
  }
}

const mapToProps = state => ({
  rules: state.getIn(['rules', 'list']),
  state: state.getIn(['rules', 'state']),
  devices: state.getIn(['devices', 'list']),
});

export default connect(mapToProps)(Automation);
