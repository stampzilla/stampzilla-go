import React, { Component } from 'react';
import { Button } from 'reactstrap';
import { connect } from 'react-redux';
import Form from 'react-jsonschema-form';

import { add, save, remove } from '../../ducks/persons';
import { write } from '../../components/Websocket';
import Card from '../../components/Card';
import CustomCheckbox from '../../components/CustomCheckbox';

const schema = {
  type: 'object',
  required: ['name'],
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    email: {
      type: 'string',
      title: 'Email',
    },
    allow_login: {
      type: 'boolean',
      title: 'Allow login',
    },
    is_admin: {
      type: 'boolean',
      title: 'Administrator',
    },
  },

  dependencies: {
    allow_login: {
      oneOf: [
        {
          properties: {
            allow_login: {
              enum: [true],
            },
            username: {
              type: 'string',
              title: 'Username',
            },
            new_password: {
              type: 'string',
              title: 'Password',
            },
            repeat_password: {
              type: 'string',
              title: 'Repeat password',
            },
          },
        },
      ],
    },
  },
};
const uiSchema = {
  'ui:order': [
    'name',
    'email',
    'allow_login',
    'username',
    'new_password',
    'repeat_password',
    '*',
  ],
  new_password: {
    'ui:widget': 'password',
  },
  repeat_password: {
    'ui:widget': 'password',
  },
};

class Person extends Component {
  state = {
    isValid: true,
    formData: {},
  };

  componentDidMount = () => {
    this.componentWillReceiveProps(this.props);
  };

  componentWillReceiveProps = (props) => {
    const { persons, match } = props;
    const person = persons && persons.find((n) => n.get('uuid') === match.params.uuid);
    if (person) {
      this.setState({
        formData: person.toJS(),
      });
    }
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
    const { persons, match, dispatch } = this.props;
    const person = persons.find((n) => n.get('uuid') === match.params.uuid);

    const postData = {
      ...(person && person.toJS()),
      ...formData,
    };

    // write({
    // type: 'update-person',
    // body: {
    // ...person.toJS(),
    // ...formData,
    // },
    // });

    if (formData.uuid) {
      console.log('save', await dispatch(save(postData)));
    } else {
      dispatch(add(postData));
    }
    this.onBackClick();
  };

  onRemove = () => {
    if (confirm('Are you sure?')) {
      const { match, dispatch } = this.props;
      dispatch(remove(match.params.uuid));
      this.onBackClick();
    }
  };

  onBackClick = () => {
    const { history } = this.props;
    history.push('/persons');
  };

  render() {
    const { persons, match } = this.props;
    const person = persons.find((n) => n.get('uuid') === match.params.uuid);

    return (
      <div className="content">
        <div className="row">
          <div className="col-md-12">
            <Card
              title={
                person ? (
                  <>
                    Settings for person
                    {' '}
                    <strong>{person.get('uuid')}</strong>
                  </>
                ) : (
                  'New person'
                )
              }
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
                <Button color="secondary" onClick={this.onBackClick}>
                  Back
                </Button>
                <Button
                  color="danger"
                  disabled={this.props.disabled}
                  onClick={this.onRemove}
                  className="ml-2 btn-sm"
                >
                  Remove
                </Button>

                <Button
                  color="primary"
                  disabled={!this.state.isValid || this.props.disabled}
                  onClick={() => this.submitButton.click()}
                  className="float-right"
                >
                  Save
                </Button>
              </div>
            </Card>
          </div>
        </div>
      </div>
    );
  }
}

const mapToProps = (state) => ({
  persons: state.getIn(['persons', 'list']),
  connections: state.getIn(['connections', 'list']),
});

export default connect(mapToProps)(Person);
