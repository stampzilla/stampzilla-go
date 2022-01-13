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
import { add, save } from '../../ducks/senders';
import Card from '../../components/Card';
import Editor from '../../components/editor';

const schema = fromJS({
  type: 'object',
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    type: {
      title: 'Type of sender',
      type: 'string',
      enum: ['file', 'email', 'webhook', 'wirepusher', 'pushbullet', 'nx', 'pushover'],
      enumNames: [
        'Logfile writer',
        'Email',
        'Webhook',
        'WirePusher',
        'Pushbullet',
        'Nx Witness Event',
        'PushOver',
      ],
    },
  },
  required: ['name', 'type'],
});
const typeSchema = {
  file: {
    schema: {
      append: {
        title: 'Append to existing file',
        type: 'boolean',
      },
      timestamp: {
        title: 'Add timestamp to each row',
        type: 'boolean',
      },
    },
  },
  email: {
    schema: {
      server: {
        title: 'Server',
        type: 'string',
      },
      port: {
        title: 'Port',
        type: 'number',
      },

      from: {
        title: 'From address',
        type: 'string',
      },
      password: {
        title: 'Password',
        type: 'string',
      },
    },
  },
  webhook: {
    schema: {
      method: {
        title: 'Method',
        type: 'string',
        enum: ['GET', 'POST', 'PUT', 'DELETE'],
      },
    },
  },
  wirepusher: {
    schema: {
      title: {
        title: 'Notification title',
        type: 'string',
        default: 'stampzilla',
      },
      type: {
        title: 'Notification type',
        type: 'string',
      },
      action: {
        title: 'Action URI',
        type: 'string',
      },
    },
    uiSchema: {
      type: {
        'ui:help':
          'Defines the ringtone, icon and vibration pattern for the notification, you can define the types on your mobile phone under "Types"',
      },
      action: {
        'ui:help':
          'URI that is going to be called when the user clicks on the notification. Could for example be an intent or a URL.',
      },
    },
  },
  pushbullet: {
    schema: {
      token: {
        title: 'API access token',
        type: 'string',
      },
    },
    uiSchema: {
      token: {
        'ui:help':
          "The Pushbullet API lets you send/receive pushes and do everything else the official Pushbullet clients can do. To access the API you'll need an access token so the server knows who you are. You can get one from your Account Settings page.",
      },
    },
  },
  pushover: {
    schema: {
      token: {
        title: 'API access token',
        type: 'string',
      },
    },
    uiSchema: {
      token: {
        'ui:help':
          'Pushover api app token',
      },
    },
  },
  nx: {
    schema: {
      server: {
        title: 'Server URL',
        type: 'string',
      },
      username: {
        title: 'Username',
        type: 'string',
      },
      password: {
        title: 'Password',
        type: 'string',
      },
    },
    uiSchema: {
      server: {
        'ui:help':
          'The ip or dns name of the NX Witness server and port in URL format. Ex http://127.0.0.1:7001/',
      },
    },
  },
};

const uiSchema = fromJS({});

const loadFromProps = (props) => {
  const { senders, match } = props;
  const rule = senders.find((n) => n.get('uuid') === match.params.uuid);
  const formData = rule && rule.toJS();

  // if (rule) {
  // formData.actions = formData.actions || [];
  // }
  return { formData };
};

class Sender extends Component {
  constructor(props) {
    super();

    this.state = {
      isValid: true,
      isModified: false,
      ...loadFromProps(props),
    };
  }

  componentWillReceiveProps(nextProps) {
    const { senders, match } = nextProps;
    if (
      !this.props
      || match.params.uuid !== this.props.match.params.uuid
      || senders !== this.props.senders
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
    const { match } = this.props;
    const { isModified, formData } = this.state;
    const { type } = formData || {};

    const patchedSchema = {
      ...schema.toJS(),
      properties: {
        ...schema.toJS().properties,
        ...(type
          && typeSchema[type] && {
          parameters: {
            title: 'Parameters',
            properties: {
              ...typeSchema[type].schema,
            },
          },
        }),
      },
    };
    const patchedUiSchema = {
      ...uiSchema.toJS(),
      parameters: {
        ...(type && typeSchema[type] && typeSchema[type].uiSchema),
      },
    };

    return (
      <>
        <div className="row">
          <div className="col-md-12">
            <Card
              title={(match.params.uuid ? 'Edit sender ' : 'New sender').concat(
                isModified ? ' (not saved)' : '',
              )}
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
          </div>
        </div>
      </>
    );
  }
}

const mapToProps = (state) => ({
  senders: state.getIn(['senders', 'list']),
});

export default connect(mapToProps)(Sender);
