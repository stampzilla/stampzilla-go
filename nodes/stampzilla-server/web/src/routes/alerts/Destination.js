import React, { Component } from 'react';
import { Button } from 'reactstrap';
import { connect } from 'react-redux';
import { fromJS } from 'immutable';
import Form from 'react-jsonschema-form';

import { request } from '../../components/Websocket';

import {
  ArrayFieldTemplate,
  ConnectedRuleConditions,
  CustomCheckbox,
  ObjectFieldTemplate,
} from '../../components/formComponents';
import { add, save } from '../../ducks/destinations';
import Card from '../../components/Card';

const titles = {
  file: 'Filename',
  email: 'Email address',
  webhook: 'Webhook url',
  wirepusher: 'WirePusher ID',
  pushbullet: 'Device identifier',
  nx: 'Associated camera',
  pushover: 'User key',
};

const schema = fromJS({
  type: 'object',
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    type: {
      title: 'Type',
      type: 'string',
      enum: ['file', 'email', 'webhook', 'wirepusher', 'pushbullet', 'nx', 'pushover'],
      enumNames: [
        'Filename',
        'Email',
        'URL',
        'WirePusher ID',
        'Pushbullet device identifier',
        'Nx Witness Generic Event',
        'PushOver user key',
      ],
    },
    sender: {},
    destinations: {
      type: 'array',
      title: 'Destinations',
      description: 'Add all destination addresses/values here',
      items: {
        type: 'string',
      },
    },
  },
  required: ['name', 'type'],
});

const uiSchema = fromJS({});

const testSchema = {
  type: 'object',
  properties: {
    body: {
      type: 'string',
      title: 'Body',
      default: 'Test message from web gui',
    },
  },
};

const testUiSchema = {};

const loadFromProps = (props) => {
  const { destinations, match } = props;
  const rule = destinations.find((n) => n.get('uuid') === match.params.uuid);
  const formData = rule && rule.toJS();

  return {
    formData: {
      ...formData,
      destinations: (formData && formData.destinations) || [],
    },
  };
};

class Destination extends Component {
  constructor(props) {
    super();

    this.state = {
      isValid: true,
      isModified: false,
      destinations: null,
      ...loadFromProps(props),
    };

    this.requestSenderDestinations(this.state.formData.sender);
  }

  componentWillReceiveProps(nextProps) {
    const { destinations, match } = nextProps;
    if (
      !this.props
      || match.params.uuid !== this.props.match.params.uuid
      || destinations !== this.props.destinations
    ) {
      const p = loadFromProps(nextProps);
      this.setState({
        ...p,
        isModified: false,
      });

      if (p.formData) {
        this.requestSenderDestinations(p.formData.sender);
      }
    }
  }

  onChange = () => (data) => {
    const { errors, formData } = data;
    this.setState({
      isValid: errors.length === 0,
      formData,
      isModified: true,
    });

    this.requestSenderDestinations(formData.sender);
  };

  requestSenderDestinations = (uuid) => {
    if (this.state.sender !== uuid) {
      request({
        type: 'sender-destinations',
        body: {
          uuid,
        },
      })
        .then((d) => this.setState({
          destinations: d,
          sender: uuid,
        }))
        .catch(() => this.setState({
          destinations: null,
          sender: uuid,
        }));
    }
  };

  onChangeTest = () => (data) => {
    const { formData } = data;
    this.setState({
      testFormData: formData,
    });
  };

  onSubmit = () => ({ formData }) => {
    const { dispatch } = this.props;

    const postData = {
      ...formData,
      uuid: this.props.match.params.uuid,
    };

    if (formData.uuid) {
      dispatch(save(postData));
    } else {
      dispatch(add(postData));
    }

    const { history } = this.props;
    history.push('/alerts');
  };

  onBackClick = () => () => {
    const { history } = this.props;
    history.push('/alerts');
  };

  onTestSubmit = (release) => ({ formData }) => {
    this.setState({
      testDisabled: true,
      triggerError: null,
      releaseError: null,
    });
    request({
      type: 'trigger-destination',
      body: {
        ...formData,
        release,
        uuid: this.props.match.params.uuid,
      },
    })
      .then(() => this.setState(({ triggerError, releaseError }) => ({
        testDisabled: false,
        triggerError: !release ? false : triggerError,
        releaseError: release ? false : releaseError,
      })))
      .catch((err) => this.setState(({ triggerError, releaseError }) => ({
        testDisabled: false,
        triggerError: !release ? err : triggerError,
        releaseError: release ? err : releaseError,
      })));
  };

  render() {
    const { match, senders } = this.props;
    const {
      isModified,
      formData,
      testFormData,
      testDisabled,
      triggerError,
      releaseError,
      destinations,
    } = this.state;
    const { sender, type } = formData || {};

    const patchedSchema = {
      ...schema.toJS(),
      // properties: {
      // },
    };
    const patchedUiSchema = {
      ...uiSchema.toJS(),
      // parameters: {},
    };

    patchedSchema.properties.destinations.title = titles[type] || 'Destinations';

    const matchingSenders = senders.filter((a) => a.get('type') === type);
    if (matchingSenders.size > 0) {
      patchedSchema.properties.sender = {
        type: 'string',
        title: 'Sender',
        enum: matchingSenders
          .map((a) => a.get('uuid'))
          .valueSeq()
          .toArray(),
        enumNames: matchingSenders
          .map((a) => a.get('name'))
          .valueSeq()
          .toArray(),
      };
    }

    if (destinations) {
      patchedSchema.properties.destinations.items.enum = Object.keys(
        destinations,
      );
      patchedSchema.properties.destinations.items.enumNames = Object.values(
        destinations,
      );
    } else {
      delete patchedSchema.properties.destinations.items.enum;
      delete patchedSchema.properties.destinations.items.enumNames;
    }

    return (
      <>
        <div className="row">
          <div className="col-md-12">
            <Card
              title={(match.params.uuid
                ? 'Edit destination'
                : 'New destination'
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
                  formData={formData}
                  onSubmit={this.onSubmit()}
                  ObjectFieldTemplate={ObjectFieldTemplate}
                  ArrayFieldTemplate={ArrayFieldTemplate}
                  widgets={{
                    CheckboxWidget: CustomCheckbox,
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

            {match.params.uuid && (
              <Card title="Test send message" bodyClassName="p-0">
                <div className="card-body">
                  <Form
                    schema={testSchema}
                    uiSchema={testUiSchema}
                    showErrorList={false}
                    liveValidate
                    onChange={this.onChangeTest()}
                    formData={testFormData}
                    onSubmit={this.onTestSubmit(false)}
                    ObjectFieldTemplate={ObjectFieldTemplate}
                    ArrayFieldTemplate={ArrayFieldTemplate}
                    widgets={{
                      CheckboxWidget: CustomCheckbox,
                    }}
                    ref={(frm) => {
                      this.testForm = frm;
                    }}
                  >
                    <button
                      ref={(btn) => {
                        this.testSubmitButton = btn;
                      }}
                      style={{ display: 'none' }}
                      type="submit"
                    />
                  </Form>
                </div>
                <div className="card-footer">
                  {triggerError && (
                    <span style={{ color: 'red' }}>
                      Failed:
                      {triggerError}
                    </span>
                  )}
                  {releaseError && (
                    <span style={{ color: 'red' }}>
                      Failed:
                      {releaseError}
                    </span>
                  )}
                  <Button
                    color={
                      releaseError
                        ? 'danger'
                        : releaseError === false
                          ? 'success'
                          : 'primary'
                    }
                    disabled={!sender || testDisabled}
                    onClick={() => this.onTestSubmit(true)({ formData: testFormData })}
                    className="float-right ml-2"
                  >
                    {'Release'}
                  </Button>
                  <Button
                    color={
                      triggerError
                        ? 'danger'
                        : triggerError === false
                          ? 'success'
                          : 'primary'
                    }
                    disabled={!sender || testDisabled}
                    onClick={() => this.testSubmitButton.click()}
                    className="float-right"
                  >
                    {'Trigger'}
                  </Button>
                </div>
              </Card>
            )}
          </div>
        </div>
      </>
    );
  }
}

const mapToProps = (state) => ({
  senders: state.getIn(['senders', 'list']),
  destinations: state.getIn(['destinations', 'list']),
});

export default connect(mapToProps)(Destination);
